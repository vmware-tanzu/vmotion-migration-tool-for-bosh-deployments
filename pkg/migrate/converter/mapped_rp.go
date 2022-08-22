/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter

import (
	"fmt"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

type MappedRP struct {
	rpMap map[string]string
}

func NewMappedResourcePool(rpMap map[string]string) *MappedRP {
	return &MappedRP{
		rpMap: rpMap,
	}
}

func (c *MappedRP) TargetResourcePool(sourceVM *vcenter.VM) (string, error) {
	targetPool, ok := c.rpMap[sourceVM.ResourcePool]
	if !ok {
		return "", fmt.Errorf("could not find a target resource pool for VM %s in resource pool %s: "+
			"ensure you add a corresponding resource pool mapping to the config file",
			sourceVM.Name, sourceVM.ResourcePool)
	}
	return targetPool, nil
}
