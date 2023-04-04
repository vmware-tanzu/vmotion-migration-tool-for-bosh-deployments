/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter

import (
	"fmt"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
	"math/rand"
	"strings"
)

const defaultResourcePoolName = "Resources"

type MappedCompute struct {
	azMappings []AZMapping
}

type AZMapping struct {
	Source AZ
	Target AZ
}

type AZ struct {
	Datacenter   string
	Cluster      string
	ResourcePool string
	Name         string
}

func (a AZ) Equals(other AZ) bool {
	return strings.EqualFold(a.Name, other.Name) &&
		strings.EqualFold(a.ResourcePool, other.ResourcePool) &&
		strings.EqualFold(a.Datacenter, other.Datacenter) &&
		strings.EqualFold(a.Cluster, other.Cluster)
}

func NewEmptyMappedCompute() *MappedCompute {
	return NewMappedCompute([]AZMapping{})
}

func NewMappedCompute(azMappings []AZMapping) *MappedCompute {
	return &MappedCompute{
		azMappings: azMappings,
	}
}

func (c *MappedCompute) TargetCompute(sourceVM *vcenter.VM) (AZ, error) {
	if sourceVM == nil {
		return AZ{}, fmt.Errorf("expected source VM to be non-nil")
	}
	az := AZ{
		Datacenter:   sourceVM.Datacenter,
		Cluster:      sourceVM.Cluster,
		ResourcePool: sourceVM.ResourcePool,
		Name:         sourceVM.AZ,
	}
	return c.TargetComputeFromSourceAZ(az)
}

func (c *MappedCompute) TargetComputeFromSourceAZ(srcAZ AZ) (AZ, error) {
	t, err := c.TargetComputesFromSourceAZ(srcAZ)
	if err != nil {
		return AZ{}, err
	}

	// pick a target at random
	return t[rand.Intn(len(t))], nil
}

func (c *MappedCompute) TargetComputesFromSourceAZ(srcCompute AZ) ([]AZ, error) {
	if srcCompute.Datacenter == "" {
		return nil, fmt.Errorf("expected datacenter to be non-empty string")
	}
	if srcCompute.Cluster == "" {
		return nil, fmt.Errorf("expected cluster to be non-empty string")
	}
	if srcCompute.Name == "" {
		return nil, fmt.Errorf("expected AZ name to be non-empty string")
	}

	if isDefaultResourcePool(srcCompute.ResourcePool) {
		srcCompute.ResourcePool = ""
	}

	// get all targets for all matching sources
	var targets []AZ
	for _, m := range c.azMappings {
		if srcCompute.Equals(m.Source) {
			targets = append(targets, m.Target)
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("could not find target compute for VM in source AZ %s, "+
			"datacenter %s, cluster %s, resource pool %s: ensure you add a corresponding compute mapping to the config file",
			srcCompute.Name, srcCompute.Datacenter, srcCompute.Cluster, srcCompute.ResourcePool)
	}
	return targets, nil
}

func (c *MappedCompute) Add(source AZ, target AZ) *MappedCompute {
	m := AZMapping{
		Source: source,
		Target: target,
	}
	c.azMappings = append(c.azMappings, m)
	return c
}

func isDefaultResourcePool(rp string) bool {
	return rp == defaultResourcePoolName
}
