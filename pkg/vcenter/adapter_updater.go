package vcenter

import (
	"context"
	"errors"

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

	l.Debugf("Using portgroup key %s", info.Port.PortgroupKey)

	backing := &types.VirtualEthernetCardDistributedVirtualPortBackingInfo{
		Port: types.DistributedVirtualSwitchPortConnection{
			PortgroupKey: info.Port.PortgroupKey,
			SwitchUuid:   info.Port.SwitchUuid,
		},
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
