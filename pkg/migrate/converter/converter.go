/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter

import (
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

type NetworkMapper interface {
	TargetNetworks(sourceVM *vcenter.VM) (map[string]string, error)
}

type ResourcePoolMapper interface {
	TargetResourcePool(sourceVM *vcenter.VM) (string, error)
}

type DatastoreMapper interface {
	TargetDatastores(sourceVM *vcenter.VM) (map[string]string, error)
}

type Converter struct {
	rpMapper         ResourcePoolMapper
	netMapper        NetworkMapper
	dsMapper         DatastoreMapper
	targetDatacenter string
	targetCluster    string
}

func New(net NetworkMapper, rp ResourcePoolMapper, ds DatastoreMapper, targetDatacenter, targetCluster string) *Converter {
	return &Converter{
		rpMapper:         rp,
		netMapper:        net,
		dsMapper:         ds,
		targetDatacenter: targetDatacenter,
		targetCluster:    targetCluster,
	}
}

func (c *Converter) TargetSpec(sourceVM *vcenter.VM) (*vcenter.TargetSpec, error) {
	rp, err := c.rpMapper.TargetResourcePool(sourceVM)
	if err != nil {
		return nil, err
	}
	nets, err := c.netMapper.TargetNetworks(sourceVM)
	if err != nil {
		return nil, err
	}
	datastores, err := c.dsMapper.TargetDatastores(sourceVM)
	if err != nil {
		return nil, err
	}

	return &vcenter.TargetSpec{
		Name:         sourceVM.Name,
		Datacenter:   c.targetDatacenter,
		Cluster:      c.targetCluster,
		ResourcePool: rp,
		Datastores:   datastores,
		Networks:     nets,
	}, nil
}
