/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import (
	"context"
	"fmt"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type Finder struct {
	Datacenter string
	client     *govmomi.Client
	finder     *find.Finder
}

func NewFinder(datacenter string, client *govmomi.Client) *Finder {
	return &Finder{
		Datacenter: datacenter,
		client:     client,
	}
}

type AdapterNotFoundError struct {
	vmName      string
	networkName string
}

func NewAdapterNotFoundError(vmName, networkName string) *AdapterNotFoundError {
	return &AdapterNotFoundError{
		vmName:      vmName,
		networkName: networkName,
	}
}

func (e *AdapterNotFoundError) Error() string {
	return fmt.Sprintf("no network interface found for VM %s on network %s", e.vmName, e.networkName)
}

func (f *Finder) HostsInCluster(ctx context.Context, clusterName string) ([]*object.HostSystem, error) {
	l := log.FromContext(ctx)
	l.Debugf("Finding hosts in cluster %s", clusterName)

	finder, err := f.getUnderlyingFinderOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	l.Debugf("Finding cluster %s", clusterName)
	destinationCluster, err := finder.ClusterComputeResource(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to find cluster %s: %w", clusterName, err)
	}

	l.Debugf("Finding hosts in cluster %s", clusterName)
	hosts, err := destinationCluster.Hosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list ESXi hosts on cluster %s: %w", clusterName, err)
	}
	return hosts, nil
}

func (f *Finder) VirtualMachine(ctx context.Context, vmName string) (*object.VirtualMachine, error) {
	log.FromContext(ctx).Debugf("Finding virtual machine %s", vmName)

	finder, err := f.getUnderlyingFinderOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	vm, err := finder.VirtualMachine(ctx, vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to find virtual machine %s: %w", vmName, err)
	}

	return vm, nil
}

func (f *Finder) ResourcePool(ctx context.Context, fullyQualifiedResourcePoolName string) (*object.ResourcePool, error) {
	log.FromContext(ctx).Debugf("Finding resource pool %s", fullyQualifiedResourcePoolName)

	finder, err := f.getUnderlyingFinderOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	resourcePool, err := finder.ResourcePool(ctx, fullyQualifiedResourcePoolName)
	if err != nil {
		return nil, fmt.Errorf("failed to find resource pool %s: %w", fullyQualifiedResourcePoolName, err)
	}
	return resourcePool, nil
}

func (f *Finder) ResourcePoolFromSpecRef(ctx context.Context, spec TargetSpec) (*types.ManagedObjectReference, error) {
	rp, err := f.ResourcePoolFromSpec(ctx, spec)
	if err != nil {
		return nil, err
	}
	r := rp.Reference()
	return &r, nil
}

func (f *Finder) ResourcePoolFromSpec(ctx context.Context, spec TargetSpec) (*object.ResourcePool, error) {
	// sanity check
	if f.Datacenter != spec.Datacenter {
		return nil, fmt.Errorf("mismatched resource pool datacenter, expected %s but got %s",
			f.Datacenter, spec.Datacenter)
	}

	longResourcePoolName := spec.FullyQualifiedResourcePool()
	return f.ResourcePool(ctx, longResourcePoolName)
}

func (f *Finder) DatastoreRef(ctx context.Context, datastoreName string) (*types.ManagedObjectReference, error) {
	ds, err := f.Datastore(ctx, datastoreName)
	if err != nil {
		return nil, err
	}
	d := ds.Reference()
	return &d, nil
}

func (f *Finder) Datastore(ctx context.Context, datastoreName string) (*object.Datastore, error) {
	log.FromContext(ctx).Debugf("Finding datastore %s", datastoreName)

	finder, err := f.getUnderlyingFinderOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	ds, err := finder.Datastore(ctx, datastoreName)
	if err != nil {
		return nil, fmt.Errorf("failed to find datastore %s: %w", datastoreName, err)
	}

	return ds, nil
}

func (f *Finder) Disks(ctx context.Context, vm *object.VirtualMachine) ([]Disk, error) {
	l := log.FromContext(ctx)
	l.Debugf("Getting VM %s datastores", vm.Name())

	finder, err := f.getUnderlyingFinderOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	devices, err := vm.Device(ctx)
	if err != nil {
		return nil, err
	}

	var disks []Disk
	for _, device := range devices {
		switch disk := device.(type) {
		case *types.VirtualDisk:
			info, ok := disk.Backing.(types.BaseVirtualDeviceFileBackingInfo)
			if !ok {
				return nil, fmt.Errorf("could not get disk %s BaseVirtualDeviceFileBackingInfo",
					disk.DeviceInfo.GetDescription().Label)
			}

			ds := info.GetVirtualDeviceFileBackingInfo().Datastore
			dsRef, err := finder.ObjectReference(ctx, ds.Reference())
			if err != nil {
				return nil, fmt.Errorf("failed to get %s datastore reference", ds.Value)
			}

			dsName := (dsRef.(*object.Datastore)).Name()
			if dsName == "" {
				return nil, fmt.Errorf("should never happen, but found an empty datastore name for %s", ds.Value)
			}

			disks = append(disks, Disk{
				ID:        device.GetVirtualDevice().Key,
				Datastore: dsName,
			})
		}
	}

	return disks, nil
}

func (f *Finder) Cluster(ctx context.Context, vm *object.VirtualMachine) (string, error) {
	l := log.FromContext(ctx)
	l.Debugf("Getting VM %s cluster", vm.Name())

	finder, err := f.getUnderlyingFinderOrCreate(ctx)
	if err != nil {
		return "", err
	}

	var o mo.VirtualMachine
	err = vm.Properties(ctx, vm.Reference(), []string{}, &o)
	if err != nil {
		return "", err
	}

	var mh mo.HostSystem
	err = vm.Properties(ctx, o.Summary.Runtime.Host.Reference(), []string{"parent"}, &mh)
	if err != nil {
		return "", err
	}

	clusterRef, err := finder.ObjectReference(ctx, mh.Parent.Reference())
	if err != nil {
		return "", fmt.Errorf("failed to get VM %s cluster reference for host %s", vm.Name(), mh.Name)
	}

	var clusterName string
	switch t := clusterRef.(type) {
	case *object.ClusterComputeResource:
		clusterName = (clusterRef.(*object.ClusterComputeResource)).Name()
	default:
		return "", fmt.Errorf("found unsupported compute type %s", t)
	}

	if clusterName == "" {
		return "", fmt.Errorf("should never happen, but found an empty cluster name for %s", clusterRef.Reference().Value)
	}

	return clusterName, nil
}

func (f *Finder) Networks(ctx context.Context, vm *object.VirtualMachine) ([]string, error) {
	l := log.FromContext(ctx)
	l.Debugf("Getting VM %s networks", vm.Name())

	finder, err := f.getUnderlyingFinderOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	var o mo.VirtualMachine
	err = vm.Properties(ctx, vm.Reference(), []string{"network"}, &o)
	if err != nil {
		return nil, err
	}

	l.Debugf("Found %d networks, getting network names", len(o.Network))
	var nets []string
	for _, net := range o.Network {
		netRef, err := finder.ObjectReference(ctx, net.Reference())
		if err != nil {
			return nil, fmt.Errorf("failed to get %s network reference", net.Value)
		}

		var netName string
		switch t := netRef.(type) {
		case *object.DistributedVirtualPortgroup:
			netName = (netRef.(*object.DistributedVirtualPortgroup)).Name()
		case *object.Network:
			netName = (netRef.(*object.Network)).Name()
		case *object.OpaqueNetwork:
			netName = (netRef.(*object.OpaqueNetwork)).Name()
		default:
			return nil, fmt.Errorf("found unsupported network type %s", t)
		}

		if netName == "" {
			return nil, fmt.Errorf("should never happen, but found an empty network name for %s", net.Value)
		}
		nets = append(nets, netName)
	}
	return nets, nil
}

func (f *Finder) AdapterBackingInfo(ctx context.Context, networkName string) (types.BaseVirtualDeviceBackingInfo, error) {
	finder, err := f.getUnderlyingFinderOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	network, err := finder.Network(ctx, networkName)
	if err != nil {
		return nil, fmt.Errorf("failed to find target network %s: %w", networkName, err)
	}

	networkBackingInfo, err := network.EthernetCardBackingInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find target ethernet card backing info for network %s: %w", networkName, err)
	}

	return networkBackingInfo, nil
}

func (f *Finder) Adapter(ctx context.Context, vmName, networkName string) (*anyAdapter, error) {
	l := log.FromContext(ctx)
	l.Debugf("Finding VM %s adapter on network %s", vmName, networkName)

	finder, err := f.getUnderlyingFinderOrCreate(ctx)
	if err != nil {
		return nil, err
	}

	network, err := finder.Network(ctx, networkName)
	if err != nil {
		return nil, fmt.Errorf("failed to find network %s: %w", networkName, err)
	}
	l.Debugf("Found network %s (%s) with path: %s",
		networkName, network.Reference().Value, network.GetInventoryPath())

	vm, err := finder.VirtualMachine(ctx, vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to find VM %s: %w", vmName, err)
	}

	virtualDeviceList, err := vm.Device(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices for VM %s: %w", vmName, err)
	}

	for _, d := range virtualDeviceList {
		switch d.(type) {
		case *types.VirtualVmxnet3:
			l.Debugf("Found a Vmxnet3 network adapter, seeing if it's attached to %s", networkName)
			var anyAdapter anyAdapter
			anyAdapter.VirtualVmxnet3 = d.(*types.VirtualVmxnet3)

			switch anyAdapter.VirtualVmxnet3.Backing.(type) {
			case *types.VirtualEthernetCardDistributedVirtualPortBackingInfo:
				info, _ := anyAdapter.VirtualVmxnet3.Backing.(*types.VirtualEthernetCardDistributedVirtualPortBackingInfo)
				if info.Port.PortgroupKey == network.Reference().Value {
					l.Debugf("Found Vmxnet3 attached to NVDS %s", networkName)
					return &anyAdapter, nil
				} else {
					l.Debugf("Vmxnet3 with distributed ethernet net id %s was not attached to %s (%s), continuing search",
						info.Port.PortgroupKey, networkName, network.Reference().Value)
				}
			case *types.VirtualEthernetCardNetworkBackingInfo:
				info, _ := anyAdapter.VirtualVmxnet3.Backing.(*types.VirtualEthernetCardNetworkBackingInfo)
				if info.Network.Value == network.Reference().Value {
					l.Debugf("Found Vmxnet3 attached to VDS %s", networkName)
					return &anyAdapter, nil
				} else {
					l.Debugf("Vmxnet3 with standard ethernet net id %s was not attached to %s (%s), continuing search",
						info.Network.Value, networkName, network.Reference().Value)
				}
			case *types.VirtualEthernetCardOpaqueNetworkBackingInfo:
				info, _ := anyAdapter.VirtualVmxnet3.Backing.(*types.VirtualEthernetCardOpaqueNetworkBackingInfo)
				if info.OpaqueNetworkId == network.Reference().Value {
					l.Debugf("Found Vmxnet3 attached to NSX-T %s", networkName)
					return &anyAdapter, nil
				} else {
					l.Debugf("Vmxnet3 with opaque net id %s was not attached to %s (%s), continuing search",
						info.OpaqueNetworkId, networkName, network.Reference().Value)
				}
			}
		case *types.VirtualE1000:
			l.Debugf("Found a E1000 network adapter, seeing if it's attached to %s", networkName)
			var anyAdapter anyAdapter
			anyAdapter.VirtualE1000 = d.(*types.VirtualE1000)

			switch anyAdapter.VirtualE1000.Backing.(type) {
			case *types.VirtualEthernetCardDistributedVirtualPortBackingInfo:
				info, _ := anyAdapter.VirtualE1000.Backing.(*types.VirtualEthernetCardDistributedVirtualPortBackingInfo)
				if info.Port.PortgroupKey == network.Reference().Value {
					l.Debugf("Found E1000 attached to NVDS %s", networkName)
					return &anyAdapter, nil
				} else {
					l.Debugf("E1000 with distributed ethernet net id %s was not attached to %s (%s), continuing search",
						info.Port.PortgroupKey, networkName, network.Reference().Value)
				}
			case *types.VirtualEthernetCardNetworkBackingInfo:
				info, _ := anyAdapter.VirtualE1000.Backing.(*types.VirtualEthernetCardNetworkBackingInfo)
				if info.Network.Value == network.Reference().Value {
					l.Debugf("Found E1000 attached to VDS %s", networkName)
					return &anyAdapter, nil
				} else {
					l.Debugf("E1000 with standard ethernet net id %s was not attached to %s (%s), continuing search",
						info.Network.Value, networkName, network.Reference().Value)
				}
			case *types.VirtualEthernetCardOpaqueNetworkBackingInfo:
				info, _ := anyAdapter.VirtualE1000.Backing.(*types.VirtualEthernetCardOpaqueNetworkBackingInfo)
				if info.OpaqueNetworkId == network.Reference().Value {
					l.Debugf("Found E1000 attached to NSX-T %s", networkName)
					return &anyAdapter, nil
				} else {
					l.Debugf("E1000 with opaque net id %s was not attached to %s (%s), continuing search",
						info.OpaqueNetworkId, networkName, network.Reference().Value)
				}
			}
		}
	}

	return nil, NewAdapterNotFoundError(vmName, networkName)
}

func (f *Finder) getUnderlyingFinderOrCreate(ctx context.Context) (*find.Finder, error) {
	if f.finder != nil {
		return f.finder, nil
	}
	finder := find.NewFinder(f.client.Client)
	log.FromContext(ctx).Debugf("Finding datacenter %s", f.Datacenter)
	destinationDataCenter, err := finder.Datacenter(ctx, f.Datacenter)
	if err != nil {
		return nil, fmt.Errorf("failed to find datacenter %s: %w", f.Datacenter, err)
	}
	finder.SetDatacenter(destinationDataCenter)

	f.finder = finder
	return f.finder, nil
}
