/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import (
	"fmt"
	"github.com/vmware/govmomi/vim25/types"
)

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

type anyNetworkBackingInfo struct {
	info types.BaseVirtualDeviceBackingInfo
}

func (i anyNetworkBackingInfo) NetworkID() string {
	switch i.info.(type) {
	case *types.VirtualEthernetCardDistributedVirtualPortBackingInfo:
		bi, _ := i.info.(*types.VirtualEthernetCardDistributedVirtualPortBackingInfo)
		return bi.Port.PortgroupKey
	case *types.VirtualEthernetCardNetworkBackingInfo:
		bi, _ := i.info.(*types.VirtualEthernetCardNetworkBackingInfo)
		return bi.Network.Value
	case *types.VirtualEthernetCardOpaqueNetworkBackingInfo:
		bi, _ := i.info.(*types.VirtualEthernetCardOpaqueNetworkBackingInfo)
		return bi.OpaqueNetworkId
	default:
		return ""
	}
}

func (i anyNetworkBackingInfo) String() string {
	return fmt.Sprintf("%v", i.info)
}

type anyAdapter struct {
	*types.VirtualVmxnet3
	*types.VirtualE1000
}

func (a anyAdapter) BackingNetworkInfo() anyNetworkBackingInfo {
	if a.VirtualE1000 != nil {
		return anyNetworkBackingInfo{
			info: a.VirtualE1000.Backing,
		}
	} else if a.VirtualVmxnet3 != nil {
		return anyNetworkBackingInfo{
			info: a.VirtualVmxnet3.Backing,
		}
	} else {
		panic("bug: no supported vNICs found")
	}
}

func (a anyAdapter) Device() types.BaseVirtualDevice {
	if a.VirtualE1000 != nil {
		return a.VirtualE1000
	} else if a.VirtualVmxnet3 != nil {
		return a.VirtualVmxnet3
	} else {
		panic("bug: no supported vNICs found")
	}
}

func (a anyAdapter) String() string {
	if a.VirtualE1000 != nil {
		return fmt.Sprintf("%s", a.VirtualE1000.MacAddress)
	} else if a.VirtualVmxnet3 != nil {
		return fmt.Sprintf("%s", a.VirtualVmxnet3.MacAddress)
	} else {
		panic("bug: no supported vNICs found")
	}
}
