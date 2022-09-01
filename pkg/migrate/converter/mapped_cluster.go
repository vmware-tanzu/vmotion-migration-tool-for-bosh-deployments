/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter

import (
	"fmt"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

type MappedCluster struct {
	clusterMap map[string]string
}

func NewMappedCluster(clusterMap map[string]string) *MappedCluster {
	return &MappedCluster{
		clusterMap: clusterMap,
	}
}

func (c *MappedCluster) TargetClusterFromSource(sourceCluster string) (string, error) {
	targetCluster, ok := c.clusterMap[sourceCluster]
	if !ok {
		return "", fmt.Errorf("could not find a target cluster for VM in source cluster %s: "+
			"ensure you add a corresponding cluster mapping to the config file",
			sourceCluster)
	}
	return targetCluster, nil
}

func (c *MappedCluster) TargetCluster(sourceVM *vcenter.VM) (string, error) {
	targetCluster, ok := c.clusterMap[sourceVM.Cluster]
	if !ok {
		return "", fmt.Errorf("could not find a target cluster for VM %s in source cluster %s: "+
			"ensure you add a corresponding cluster mapping to the config file",
			sourceVM.Name, sourceVM.ResourcePool)
	}
	return targetCluster, nil
}
