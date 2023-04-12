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

type DatastoreMapper interface {
	TargetDatastores(sourceVM *vcenter.VM) (map[string]string, error)
}

type ComputeMapper interface {
	TargetCompute(sourceVM *vcenter.VM) (AZ, error)
}

type Converter struct {
	netMapper     NetworkMapper
	dsMapper      DatastoreMapper
	computeMapper ComputeMapper
}

func New(net NetworkMapper, ds DatastoreMapper, cm ComputeMapper) *Converter {
	return &Converter{
		netMapper:     net,
		dsMapper:      ds,
		computeMapper: cm,
	}
}

func (c *Converter) TargetSpec(sourceVM *vcenter.VM) (*vcenter.TargetSpec, error) {
	nets, err := c.netMapper.TargetNetworks(sourceVM)
	if err != nil {
		return nil, err
	}
	datastores, err := c.dsMapper.TargetDatastores(sourceVM)
	if err != nil {
		return nil, err
	}
	compute, err := c.computeMapper.TargetCompute(sourceVM)
	if err != nil {
		return nil, err
	}
	targetFolder, err := TargetFolder(sourceVM.Folder, compute.Datacenter)
	if err != nil {
		return nil, err
	}

	return &vcenter.TargetSpec{
		Name:         sourceVM.Name,
		Datacenter:   compute.Datacenter,
		Cluster:      compute.Cluster,
		ResourcePool: compute.ResourcePool,
		Folder:       targetFolder,
		Datastores:   datastores,
		Networks:     nets,
	}, nil
}
