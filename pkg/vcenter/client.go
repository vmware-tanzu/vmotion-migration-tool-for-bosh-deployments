/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import (
	"context"
	"errors"
	"fmt"
	"github.com/vmware/govmomi/find"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/thumbprint"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/session/keepalive"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/soap"
)

const keepaliveInterval = 5 * time.Minute // vCenter APIs keep-alive

type Client struct {
	host       string
	user       string
	password   string
	insecure   bool
	datacenter string

	certThumb  string
	client     *govmomi.Client
	clientOnce sync.Once
	thumbOnce  sync.Once
	initErr    error
}

func NewFromGovmomiClient(client *govmomi.Client, datacenter string) *Client {
	p, _ := client.URL().User.Password()
	return &Client{
		host:       client.URL().Host,
		user:       client.URL().User.Username(),
		password:   p,
		datacenter: datacenter,
		insecure:   false,
		client:     client,
	}
}

func New(host, username, password, datacenter string, insecure bool) *Client {
	return &Client{
		host:       host,
		user:       username,
		password:   password,
		datacenter: datacenter,
		insecure:   insecure,
	}
}

func (c *Client) UserName() string {
	return c.user
}

func (c *Client) Password() string {
	return c.password
}

func (c *Client) HostName() string {
	return c.host
}

func (c *Client) Datacenter() string {
	return c.datacenter
}

func (c *Client) Insecure() bool {
	return c.insecure
}

func (c *Client) FindVMInClusters(ctx context.Context, azName, vmNameOrPath string, clusters []string) (*VM, error) {
	vm, err := c.findVM(ctx, azName, vmNameOrPath)
	if err != nil {
		return nil, err
	}

	found := false
	for _, cl := range clusters {
		if strings.EqualFold(vm.Cluster, cl) {
			found = true
		}
	}
	if !found {
		return nil, NewVMNotFoundError(vmNameOrPath,
			fmt.Errorf("VM exists, but not in clusters %s", strings.Join(clusters, ", ")))
	}

	return vm, err
}

// CreateFolder creates the specified folder including any parent folders
// If the folder(s) already exists, no folders are created and nil is returned
func (c *Client) CreateFolder(ctx context.Context, folderPath string) error {
	l := log.FromContext(ctx)
	l.Debugf("Creating folder %s", folderPath)

	client, err := c.getOrCreateUnderlyingClient(ctx)
	if err != nil {
		return err
	}
	finder := NewFinder(c.Datacenter(), client)

	// split the path into it's base (/datacenter/vm|host|storage|network/) and subpaths
	splitFn := func(c rune) bool {
		return c == '/'
	}
	folderParts := strings.FieldsFunc(folderPath, splitFn)
	if len(folderParts) < 2 {
		return fmt.Errorf("expected a folder path with at least 2 base parts, but got %d", len(folderPath))
	}

	// get the base path /dc/vm and sub-path parts
	curFolderPath := "/" + strings.Join(folderParts[:2], "/")
	subPaths := folderParts[2:]

	// get the base path folder
	curFolder, err := finder.Folder(ctx, curFolderPath)
	if err != nil {
		return fmt.Errorf("could not find base folder '%s': %w", curFolderPath, err)
	}

	// walk each sub-folder
	for _, p := range subPaths {
		nextFolderPath := curFolderPath + "/" + p
		nextFolder, err := finder.Folder(ctx, nextFolderPath)
		if err != nil {
			var notFoundErr *find.NotFoundError
			if errors.As(err, &notFoundErr) {
				// folder does not exist create it
				nextFolder, err = curFolder.CreateFolder(ctx, p)
				if err != nil {
					if strings.Contains(err.Error(), "already exists") {
						// likely another worker _just_ created this folder - see if we can _now_ find it
						nextFolder, err = finder.Folder(ctx, nextFolderPath)
						if err != nil {
							return fmt.Errorf("folder '%s' already exists, but can't find it: %w", nextFolderPath, err)
						}
						l.Debugf("Another worker already created '%s', continuing", nextFolderPath)
					} else {
						return fmt.Errorf("could not create new sub-folder '%s': %w", nextFolderPath, err)
					}
				}
			} else {
				return fmt.Errorf("could not find the target folder '%s': %w", nextFolderPath, err)
			}
		}
		curFolderPath = nextFolderPath
		curFolder = nextFolder
	}

	return nil
}

func (c *Client) Logout(ctx context.Context) {
	if c.client != nil {
		err := c.client.Logout(ctx)
		if err != nil {
			log.FromContext(ctx).Warnf("vSphere logout failed: %s", err)
		}
	}
}

func (c *Client) URL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   c.HostName(),
		Path:   "/sdk",
	}
}

func (c *Client) isSameVCenter(host, username, password string, insecure bool) bool {
	return c.host == host && c.user == username && c.password == password && c.insecure == insecure
}

func (c *Client) findVM(ctx context.Context, azName, vmNameOrPath string) (*VM, error) {
	l := log.FromContext(ctx)

	client, err := c.getOrCreateUnderlyingClient(ctx)
	if err != nil {
		return nil, err
	}

	f := NewFinder(c.Datacenter(), client)
	vm, err := f.VirtualMachine(ctx, vmNameOrPath)
	if err != nil {
		if strings.Contains(err.Error(), "failed to find virtual machine") {
			return nil, NewVMNotFoundError(vmNameOrPath, err)
		}
		return nil, err
	}

	l.Debugf("Getting VM %s resource pool", vmNameOrPath)
	rp, err := vm.ResourcePool(ctx)
	if err != nil {
		return nil, err
	}

	cluster, err := f.VMClusterName(ctx, vm)
	if err != nil {
		return nil, err
	}

	pool, err := rp.ObjectName(ctx)
	if err != nil {
		return nil, err
	}

	nets, err := f.Networks(ctx, vm)
	if err != nil {
		return nil, err
	}

	disks, err := f.Disks(ctx, vm)
	if err != nil {
		return nil, err
	}

	return &VM{
		Name:         vm.Name(),
		AZ:           azName,
		Datacenter:   c.Datacenter(),
		Cluster:      cluster,
		ResourcePool: pool,
		Folder:       path.Dir(vm.InventoryPath),
		Networks:     nets,
		Disks:        disks,
	}, nil
}

func (c *Client) urlWithUser() *url.URL {
	u := c.URL()
	u.User = url.UserPassword(c.user, c.password)
	return u
}

func (c *Client) getOrCreateUnderlyingClient(ctx context.Context) (*govmomi.Client, error) {
	c.clientOnce.Do(func() {
		// in case already pre-populated from an external source
		if c.client != nil {
			return
		}

		l := log.FromContext(ctx)

		u := c.urlWithUser()
		l.Debugf("Creating govmomi client: %+v", u)

		soapClient := soap.NewClient(u, c.insecure)
		vimClient, err := vim25.NewClient(ctx, soapClient)
		if err != nil {
			c.initErr = fmt.Errorf("could not create new vim25 govmomi client: %w", err)
			return
		}
		vimClient.RoundTripper = keepalive.NewHandlerSOAP(
			vimClient.RoundTripper, keepaliveInterval, soapKeepAliveHandler(ctx, vimClient))

		l.Debug("Creating vim client session manager and logging in to activate keep-alive handler")
		m := session.NewManager(vimClient)
		err = m.Login(ctx, u.User)
		if err != nil {
			c.initErr = fmt.Errorf("could not login via vim25 session manager: %w", err)
			return
		}

		c.client = &govmomi.Client{
			Client:         vimClient,
			SessionManager: m,
		}
	})

	// check for a previous init err
	if c.initErr != nil {
		return nil, c.initErr
	}

	return c.client, nil
}

func soapKeepAliveHandler(ctx context.Context, c *vim25.Client) func() error {
	return func() error {
		l := log.FromContext(ctx)
		l.Infof("Executing SOAP keep-alive handler %s", c.URL())
		t, err := methods.GetCurrentTime(ctx, c)
		if err != nil {
			return err
		}

		l.Debugf("vCenter %s current time: %s", c.URL(), t.String())
		return nil
	}
}

func (c *Client) thumbprint(ctx context.Context) (string, error) {
	var initErr error
	c.thumbOnce.Do(func() {
		thumb, err := thumbprint.RetrieveSHA1(c.host, 443)
		if err != nil {
			initErr = fmt.Errorf("failed to get %s:%d cert thumbprint", c.host, 443)
			return
		}
		log.FromContext(ctx).Debugf("Target %s:%d cert thumbprint is: %s", c.host, 443, thumb)
		c.certThumb = thumb
	})

	return c.certThumb, initErr
}
