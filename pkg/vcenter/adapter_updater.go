/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import (
	"context"
	"errors"
	"fmt"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware/govmomi/vim25/types"
)

type AdapterUpdater struct {
	finder *Finder
}

func NewAdapterUpdater(finder *Finder) *AdapterUpdater {
	return &AdapterUpdater{
		finder: finder,
	}
}

func (a *AdapterUpdater) TargetNewNetwork(ctx context.Context, adapter *anyAdapter, targetNetName string) (*anyAdapter, error) {
	l := log.FromContext(ctx)
	l.Debugf("Finding target network %s", targetNetName)

	info, err := a.finder.AdapterBackingInfo(ctx, targetNetName)
	if err != nil {
		return nil, err
	}

	var backing types.BaseVirtualDeviceBackingInfo
	switch t := info.(type) {
	case *types.VirtualEthernetCardDistributedVirtualPortBackingInfo:
		bi, _ := info.(*types.VirtualEthernetCardDistributedVirtualPortBackingInfo)
		l.Debugf("Using distributed vswitch portgroup key %s", bi.Port.PortgroupKey)
		backing = &types.VirtualEthernetCardDistributedVirtualPortBackingInfo{
			Port: types.DistributedVirtualSwitchPortConnection{
				PortgroupKey: bi.Port.PortgroupKey,
				SwitchUuid:   bi.Port.SwitchUuid,
			},
		}
	case *types.VirtualEthernetCardNetworkBackingInfo:
		bi, _ := info.(*types.VirtualEthernetCardNetworkBackingInfo)
		l.Debugf("Using standard network name %s", bi.VirtualDeviceDeviceBackingInfo.DeviceName)
		backing = &types.VirtualEthernetCardNetworkBackingInfo{
			VirtualDeviceDeviceBackingInfo: types.VirtualDeviceDeviceBackingInfo{
				DeviceName: bi.VirtualDeviceDeviceBackingInfo.DeviceName,
			},
		}
	case *types.VirtualEthernetCardOpaqueNetworkBackingInfo:
		bi, _ := info.(*types.VirtualEthernetCardOpaqueNetworkBackingInfo)
		l.Debugf("Using opaque network ID %s", bi.OpaqueNetworkId)
		backing = &types.VirtualEthernetCardOpaqueNetworkBackingInfo{
			OpaqueNetworkId:   bi.OpaqueNetworkId,
			OpaqueNetworkType: bi.OpaqueNetworkType,
		}
	default:
		return nil, fmt.Errorf("unexpected network card backing info type %s", t)
	}

	if adapter.VirtualE1000 != nil {
		adapter.VirtualE1000.Backing = backing
	} else if adapter.VirtualVmxnet3 != nil {
		adapter.VirtualVmxnet3.Backing = backing
	} else {
		return nil, errors.New("bug: no supported vNICs found")
	}

	return adapter, nil
}
