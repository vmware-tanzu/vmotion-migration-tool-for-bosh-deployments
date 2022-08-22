package vcenter

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/types"
	"net/url"
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
	Host       string
	Username   string
	Password   string
	Insecure   bool
	DryRun     bool
	certThumb  string
	client     *govmomi.Client
	clientOnce sync.Once
	thumbOnce  sync.Once
}

func NewFromGovmomiClient(client *govmomi.Client) *Client {
	p, _ := client.URL().User.Password()
	return &Client{
		Host:     client.URL().Host,
		Username: client.URL().User.Username(),
		Password: p,
		Insecure: false,
		client:   client,
	}
}

func New(host, username, password string, insecure bool) *Client {
	return &Client{
		Host:     host,
		Username: username,
		Password: password,
		Insecure: insecure,
	}
}

func (c *Client) HostName() string {
	return c.Host
}

func (c *Client) FindVM(ctx context.Context, datacenter, cluster, vmName string) (*VM, error) {
	client, err := c.getOrCreateUnderlyingClient(ctx)
	if err != nil {
		return nil, err
	}

	f := NewFinder(datacenter, client)
	vm, err := f.VirtualMachine(ctx, vmName)
	if err != nil {
		if strings.Contains(err.Error(), "failed to find virtual machine") {
			return nil, NewVMNotFoundError(vmName)
		}
		return nil, err
	}

	log.FromContext(ctx).Debugf("Getting VM %s resource pool", vmName)
	rp, err := vm.ResourcePool(ctx)
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

	return &VM{
		Name:         vmName,
		Datacenter:   datacenter,
		Cluster:      cluster,
		ResourcePool: pool,
		Networks:     nets,
	}, nil
}

func (c *Client) SetEVCMode(ctx context.Context, datacenter, vmName, evcMode string) error {
	client, err := c.getOrCreateUnderlyingClient(ctx)
	if err != nil {
		return err
	}

	// find the VM
	f := NewFinder(datacenter, client)
	vm, err := f.VirtualMachine(ctx, vmName)
	if err != nil {
		if strings.Contains(err.Error(), "failed to find virtual machine") {
			return NewVMNotFoundError(vmName)
		}
		return err
	}

	// make sure it's powered off
	ps, err := vm.PowerState(ctx)
	if err != nil {
		return err
	}
	if ps != types.VirtualMachinePowerStatePoweredOff {
		return fmt.Errorf("error virtual machine %s is not powered off", vmName)
	}

	pc := property.DefaultCollector(c.client.Client)
	//obj, err := find.NewFinder(c.client.Client).VirtualMachine(ctx, "DC0_H0_VM0")
	obj, err := find.NewFinder(c.client.Client).ClusterComputeResourceOrDefault(ctx, "DC0_H0_VM0")
	pc.RetrieveOne()

	// How do I build a complete mask of all supported features for the given EVC mode?
	masks := []types.HostFeatureMask{
		{
			Key:         "",
			FeatureName: "",
			Value:       "",
		},
	}

	isComplete := true
	req := types.ApplyEvcModeVM_Task{
		This:          vm.Reference(),
		Mask:          masks,
		CompleteMasks: &isComplete,
	}

	// apply the EVC mode to the VM
	res, err := methods.ApplyEvcModeVM_Task(ctx, c.client.Client, &req)
	if err != nil {
		return err
	}

	t := object.NewTask(c.client.Client, res.Returnval)
	return t.Wait(ctx)
}

func (c *Client) URL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   c.Host,
		Path:   "/sdk",
	}
}

func (c *Client) urlWithUser() *url.URL {
	u := c.URL()
	u.User = url.UserPassword(c.Username, c.Password)
	return u
}

func (c *Client) Logout(ctx context.Context) {
	if c.client != nil {
		err := c.client.Logout(ctx)
		if err != nil {
			log.FromContext(ctx).Warnf("vSphere logout failed: %s", err)
		}
	}
}

func (c *Client) getOrCreateUnderlyingClient(ctx context.Context) (*govmomi.Client, error) {
	var initErr error
	c.clientOnce.Do(func() {
		// in case already pre-populated from an external source
		if c.client != nil {
			return
		}

		l := log.FromContext(ctx)

		u := c.urlWithUser()
		l.Debugf("Creating govmomi client: %+v", u)

		soapClient := soap.NewClient(u, c.Insecure)
		vimClient, err := vim25.NewClient(ctx, soapClient)
		if err != nil {
			initErr = fmt.Errorf("could not create new vim25 govmomi client: %w", err)
			return
		}
		vimClient.RoundTripper = keepalive.NewHandlerSOAP(
			vimClient.RoundTripper, keepaliveInterval, soapKeepAliveHandler(ctx, vimClient))

		l.Debug("Creating vim client session manager and logging in to activate keep-alive handler")
		m := session.NewManager(vimClient)
		err = m.Login(ctx, u.User)
		if err != nil {
			initErr = fmt.Errorf("could not login via vim25 session manager: %w", err)
			return
		}

		c.client = &govmomi.Client{
			Client:         vimClient,
			SessionManager: m,
		}
	})

	if initErr != nil {
		return nil, initErr
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
		thumb, err := thumbprint.RetrieveSHA1(c.Host, 443)
		if err != nil {
			initErr = fmt.Errorf("failed to get %s:%d cert thumbprint", c.Host, 443)
			return
		}
		log.FromContext(ctx).Debugf("Target %s:%d cert thumbprint is: %s", c.Host, 443, thumb)
		c.certThumb = thumb
	})

	return c.certThumb, initErr
}
