/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
)

type hostRef struct {
	host       *object.HostSystem
	leaseCount int
}

type HostPool struct {
	Datacenter string

	MaxLeasePerHost             int
	LeaseWaitTimeoutInMinutes   int
	LeaseCheckIntervalInSeconds int

	client         *Client
	clusterToHosts map[string][]*hostRef
	initialized    bool
}

func NewHostPool(client *Client, datacenter string) *HostPool {
	return &HostPool{
		client:          client,
		Datacenter:      datacenter,
		clusterToHosts:  make(map[string][]*hostRef),
		MaxLeasePerHost: 1, // padded down to 1 instead of 2 - could be much higher w/o storage vmotion
		// https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vcenterhost.doc/GUID-25EA5833-03B5-4EDD-A167-87578B8009B3.html
		LeaseWaitTimeoutInMinutes:   20,
		LeaseCheckIntervalInSeconds: 10,
	}
}

func (hp *HostPool) Initialize(ctx context.Context) error {
	l := log.FromContext(ctx)
	if hp.initialized {
		l.Debug("Host pool already initialized, skipping init...")
		return nil
	}

	l.Debugf("Initializing host pool for datacenter %s", hp.Datacenter)

	c, err := hp.client.getOrCreateUnderlyingClient(ctx)
	if err != nil {
		return err
	}

	finder := find.NewFinder(c.Client)
	l.Debugf("Finding datacenter %s", hp.Datacenter)
	destinationDataCenter, err := finder.Datacenter(ctx, hp.Datacenter)
	if err != nil {
		return fmt.Errorf("failed to find datacenter %s: %w", hp.Datacenter, err)
	}
	finder.SetDatacenter(destinationDataCenter)

	dcPath := fmt.Sprintf("/%s/host/*", hp.Datacenter)
	l.Debugf("Listing clusters in %s", dcPath)
	clusters, err := finder.ClusterComputeResourceList(ctx, dcPath)
	if err != nil {
		return fmt.Errorf("failed to list clusters on datacenter %s: %w", hp.Datacenter, err)
	}

	for _, cluster := range clusters {
		l.Debugf("Listing ESXi hosts in cluster %s", cluster.Name())
		hosts, err := cluster.Hosts(ctx)
		if err != nil {
			return fmt.Errorf("failed to list ESXi hosts on cluster %s: %w", cluster.Name(), err)
		}

		// find all the hosts in the cluster that are not in maintenance mode
		for _, h := range hosts {
			var hmo mo.HostSystem
			err = h.Properties(ctx, h.Reference(), []string{"runtime"}, &hmo)
			if err != nil {
				return fmt.Errorf("could not get runtime properties for host %s: %w", h.Name(), err)
			}
			if hmo.Runtime.InMaintenanceMode {
				l.Debugf("Found host %s in maintenance mode, ignoring", h.Name())
				continue
			}
			l.Debugf("Adding ESXi host %s to host pool", h.Name())
			hosts := hp.clusterToHosts[cluster.Name()]
			hosts = append(hosts, &hostRef{
				host: h,
			})
			hp.clusterToHosts[cluster.Name()] = hosts
		}
	}

	hp.initialized = true
	return nil
}

// LeaseAvailableHost returns the best host system to copy a VM to
// If no hosts are currently available a nil host will be returned, the caller should wait and retry later
// Release should be called by the caller when done with the host
func (hp *HostPool) LeaseAvailableHost(ctx context.Context, clusterName string) (*object.HostSystem, error) {
	// TODO this has to be synchronized
	l := log.FromContext(ctx)
	if !hp.initialized {
		return nil, fmt.Errorf("host pool not initialized")
	}

	hostRefs, ok := hp.clusterToHosts[clusterName]
	if !ok {
		return nil, fmt.Errorf("found no hosts on cluster %s", clusterName)
	}

	// find all hosts with the least amount of leases
	// if none found with zero leases, try hosts with 1 and so on up to max
	var hostCandidates []*hostRef
	for optimalLeaseCount := 0; optimalLeaseCount < hp.MaxLeasePerHost; optimalLeaseCount++ {
		for _, r := range hostRefs {
			if r.leaseCount == optimalLeaseCount {
				l.Debugf("Found candidate host %s on cluster %s with %d leases", r.host.Name(), clusterName, optimalLeaseCount)
				hostCandidates = append(hostCandidates, r)
			}
		}
		if len(hostCandidates) > 0 {
			break
		}
	}

	if len(hostCandidates) == 0 {
		// no hosts exist with less than max lease per host
		l.Debugf("Found no hosts on cluster %s with %d or fewer leases", clusterName, hp.MaxLeasePerHost)
		return nil, nil
	}

	// pick a random host from list of available hosts, avoid always picking first host in list
	h := hostCandidates[rand.Intn(len(hostCandidates))]
	h.leaseCount++
	return h.host, nil
}

// WaitForLeaseAvailableHost returns the best host system to copy a VM to
// If no hosts are currently available this func will block until one is available or configured timeout
// Release should be called by the caller when done with the host
func (hp *HostPool) WaitForLeaseAvailableHost(ctx context.Context, clusterName string) (*object.HostSystem, error) {
	timeout := time.After(time.Minute * time.Duration(hp.LeaseWaitTimeoutInMinutes))
	ticker := time.NewTicker(time.Second * time.Duration(hp.LeaseCheckIntervalInSeconds))
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("unable to find a target host on cluster %s after %d minutes, giving up",
				clusterName, hp.LeaseWaitTimeoutInMinutes)
		case <-ticker.C:
			targetHost, err := hp.LeaseAvailableHost(ctx, clusterName)
			if err != nil {
				return nil, err
			}
			if targetHost != nil {
				return targetHost, nil
			}
		}
	}
}

// Release releases the specified host back into the pool and makes it available for lease again
func (hp *HostPool) Release(ctx context.Context, host *object.HostSystem) {
	if host == nil {
		return
	}

	// this is pretty brute force...
	l := log.FromContext(ctx)
	for _, hostRefs := range hp.clusterToHosts {
		for _, hRef := range hostRefs {
			if hRef.host == host {
				l.Debugf("Releasing lease on host %s", host.Name())
				hRef.leaseCount--
				return
			}
		}
	}
	l.Warnf("Could not find lease on host %s, is there a ref leak?", host.Name())
}
