/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter

import (
	"fmt"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

type MappedNet struct {
	networkMap map[string]string
}

func NewEmptyMappedNetwork() *MappedNet {
	return NewMappedNetwork(map[string]string{})
}

func NewMappedNetwork(networkMap map[string]string) *MappedNet {
	return &MappedNet{
		networkMap: networkMap,
	}
}

func (m *MappedNet) Add(srcNet, targetNet string) *MappedNet {
	m.networkMap[srcNet] = targetNet
	return m
}

func (m *MappedNet) TargetNetworks(sourceVM *vcenter.VM) (map[string]string, error) {
	targetNetworks := map[string]string{}
	for _, src := range sourceVM.Networks {
		target, ok := m.networkMap[src]
		if !ok {
			return nil, fmt.Errorf("could not find a target network for VM %s attached to network %s: "+
				"ensure you add a corresponding network mapping to the config file",
				sourceVM.Name, src)
		}
		targetNetworks[src] = target
	}
	return targetNetworks, nil
}
