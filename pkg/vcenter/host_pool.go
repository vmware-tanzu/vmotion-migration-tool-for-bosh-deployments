/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
)

type hostRef struct {
	host         *object.HostSystem
	leaseCount   int
	leaseStart   time.Time
	leaseRelease time.Time
}

func (hr *hostRef) StartLease() {
	hr.leaseStart = time.Now()
	hr.leaseCount++
}

func (hr *hostRef) ReleaseLease() {
	hr.leaseRelease = time.Now()
	hr.leaseCount--
}

type HostPoolConfig struct {
	AZs map[string]HostPoolAZ
}

type HostPoolAZ struct {
	Clusters []string
}

type HostPool struct {
	MaxLeasePerHost             int
	LeaseWaitTimeoutInMinutes   int
	LeaseCheckIntervalInSeconds int

	clientPool  *Pool
	config      *HostPoolConfig
	azToHosts   map[string][]*hostRef
	initialized bool
	leaseMutex  sync.Mutex
}

func NewHostPool(clientPool *Pool, config *HostPoolConfig) *HostPool {
	return &HostPool{
		clientPool:      clientPool,
		config:          config,
		azToHosts:       make(map[string][]*hostRef),
		MaxLeasePerHost: 1, // padded down to 1 instead of 2 - could be much higher w/o storage vmotion
		// https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vcenterhost.doc/GUID-25EA5833-03B5-4EDD-A167-87578B8009B3.html
		LeaseWaitTimeoutInMinutes:   30,
		LeaseCheckIntervalInSeconds: 30,
	}
}

func (hp *HostPool) Initialize(ctx context.Context) error {
	l := log.FromContext(ctx)
	if hp.initialized {
		l.Debug("Host pool already initialized, skipping init...")
		return nil
	}

	for n, az := range hp.config.AZs {
		client := hp.clientPool.GetTargetClientByAZ(n)
		err := hp.initializeHostPoolForAZ(ctx, n, az.Clusters, client)
		if err != nil {
			return err
		}
	}

	hp.initialized = true
	return nil
}

// LeaseAvailableHost returns the best host system to copy a VM to
// If no hosts are currently available a nil host will be returned, the caller should wait and retry later
// Release should be called by the caller when done with the host
func (hp *HostPool) LeaseAvailableHost(ctx context.Context, azName string) (*object.HostSystem, error) {
	hp.leaseMutex.Lock()
	defer hp.leaseMutex.Unlock()

	l := log.FromContext(ctx)
	if !hp.initialized {
		return nil, fmt.Errorf("host pool not initialized")
	}

	hostRefs, ok := hp.azToHosts[azName]
	if !ok {
		return nil, fmt.Errorf("found no hosts in az %s", azName)
	}

	// find all hosts with the least amount of leases
	// if none found with zero leases, try hosts with 1 and so on up to max
	var hostCandidates []*hostRef
	for optimalLeaseCount := 0; optimalLeaseCount < hp.MaxLeasePerHost; optimalLeaseCount++ {
		for _, r := range hostRefs {
			if r.leaseCount == optimalLeaseCount {
				l.Debugf("Found candidate host %s on az %s with %d leases", r.host.Name(), azName, optimalLeaseCount)
				hostCandidates = append(hostCandidates, r)
			}
		}
		if len(hostCandidates) > 0 {
			break
		}
	}

	if len(hostCandidates) == 0 {
		// no hosts exist with less than max lease per host
		l.Debugf("Found no hosts on az %s with %d or fewer leases", azName, hp.MaxLeasePerHost)
		return nil, nil
	}

	// sort candidate hosts by last lease time
	// pick the host that has the most time since last lease
	sort.Slice(hostCandidates, func(i, j int) bool {
		return hostCandidates[i].leaseRelease.Before(hostCandidates[j].leaseRelease)
	})
	h := hostCandidates[0]
	h.StartLease()
	return h.host, nil
}

// WaitForLeaseAvailableHost returns the best host system to copy a VM to
// If no hosts are currently available this func will block until one is available or configured timeout
// Release should be called by the caller when done with the host
func (hp *HostPool) WaitForLeaseAvailableHost(ctx context.Context, azName string) (*object.HostSystem, error) {
	timeout := time.After(time.Minute * time.Duration(hp.LeaseWaitTimeoutInMinutes))
	ticker := time.NewTicker(time.Second * time.Duration(hp.LeaseCheckIntervalInSeconds))
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("unable to find a target host on az %s after %d minutes, giving up",
				azName, hp.LeaseWaitTimeoutInMinutes)
		case <-ticker.C:
			targetHost, err := hp.LeaseAvailableHost(ctx, azName)
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

	hp.leaseMutex.Lock()
	defer hp.leaseMutex.Unlock()

	// this is pretty brute force...
	l := log.FromContext(ctx)
	for _, hostRefs := range hp.azToHosts {
		for _, hRef := range hostRefs {
			if hRef.host == host {
				l.Debugf("Releasing lease on host %s", host.Name())
				hRef.ReleaseLease()
				return
			}
		}
	}
	l.Warnf("Could not find lease on host %s, is there a ref leak?", host.Name())
}

func (hp *HostPool) initializeHostPoolForAZ(ctx context.Context, az string, clusterNames []string, client *Client) error {
	l := log.FromContext(ctx)
	l.Debugf("Initializing host pool for az %s", az)

	c, err := client.getOrCreateUnderlyingClient(ctx)
	if err != nil {
		return err
	}

	finder := find.NewFinder(c.Client)
	l.Debugf("Finding az %s datacenter %s", az, client.Datacenter())
	destinationDataCenter, err := finder.Datacenter(ctx, client.Datacenter())
	if err != nil {
		return fmt.Errorf("failed to find az %s datacenter %s: %w", az, client.Datacenter(), err)
	}
	finder.SetDatacenter(destinationDataCenter)

	// only include the clusters explicitly listed in the config
	var clusters []*object.ClusterComputeResource
	for _, n := range clusterNames {
		cc, err := finder.ClusterComputeResource(ctx, n)
		if err != nil {
			return fmt.Errorf("failed to get az %s cluster %s on datacenter %s: %w", az, n, client.Datacenter(), err)
		}
		clusters = append(clusters, cc)
	}

	// list out all the hosts for each cluster
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
			azHosts := hp.azToHosts[az]
			azHosts = append(azHosts, &hostRef{
				host: h,
			})
			hp.azToHosts[az] = azHosts
		}
	}

	return nil
}
