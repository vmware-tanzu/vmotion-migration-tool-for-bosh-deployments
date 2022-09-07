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

type ClusterMapper interface {
	TargetCluster(sourceVM *vcenter.VM) (string, error)
}

type Converter struct {
	rpMapper         ResourcePoolMapper
	netMapper        NetworkMapper
	dsMapper         DatastoreMapper
	clusterMapper    ClusterMapper
	targetDatacenter string
}

func New(net NetworkMapper, rp ResourcePoolMapper, ds DatastoreMapper, cm ClusterMapper, targetDatacenter string) *Converter {
	return &Converter{
		rpMapper:         rp,
		netMapper:        net,
		dsMapper:         ds,
		clusterMapper:    cm,
		targetDatacenter: targetDatacenter,
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
	cluster, err := c.clusterMapper.TargetCluster(sourceVM)
	if err != nil {
		return nil, err
	}

	return &vcenter.TargetSpec{
		Name:         sourceVM.Name,
		Datacenter:   c.targetDatacenter,
		Cluster:      cluster,
		ResourcePool: rp,
		Datastores:   datastores,
		Networks:     nets,
	}, nil
}

func (c *Converter) TargetDatacenter() string {
	return c.targetDatacenter
}
