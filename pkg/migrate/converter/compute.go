/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter

import (
	"fmt"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

const defaultResourcePoolName = "Resources"

type MappedCompute struct {
	srcToDstCompute map[AZMapping]AZMapping
}

type AZMapping struct {
	Cluster      string
	ResourcePool string
	VCenterHost  string
}

func NewEmptyMappedCompute() *MappedCompute {
	return NewMappedCompute(map[AZMapping]AZMapping{})
}

func NewMappedCompute(computeMap map[AZMapping]AZMapping) *MappedCompute {
	return &MappedCompute{
		srcToDstCompute: computeMap,
	}
}

func (c *MappedCompute) TargetCompute(sourceVM *vcenter.VM) (AZMapping, error) {
	az := AZMapping{
		Cluster:      sourceVM.Cluster,
		ResourcePool: sourceVM.ResourcePool,
	}
	return c.TargetComputeFromSource(az)
}

func (c *MappedCompute) TargetComputeFromSource(srcCompute AZMapping) (AZMapping, error) {
	if isDefaultResourcePool(srcCompute.ResourcePool) {
		srcCompute.ResourcePool = ""
	}
	dstAZMapping, ok := c.srcToDstCompute[srcCompute]
	if !ok {
		return AZMapping{}, fmt.Errorf("could not find target compute for VM in source "+
			"cluster %s, resource pool %s: ensure you add a corresponding compute mapping to the config file",
			srcCompute.Cluster, srcCompute.ResourcePool)
	}
	return dstAZMapping, nil
}

func (c *MappedCompute) Add(src AZMapping, dst AZMapping) *MappedCompute {
	c.srcToDstCompute[src] = dst
	return c
}

func isDefaultResourcePool(rp string) bool {
	return rp == defaultResourcePoolName
}
