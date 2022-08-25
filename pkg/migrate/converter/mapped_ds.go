/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter

import (
	"fmt"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

type MappedDS struct {
	dsMap map[string]string
}

func NewEmptyMappedDatastore() *MappedDS {
	return NewMappedDatastore(map[string]string{})
}

func NewMappedDatastore(dsMap map[string]string) *MappedDS {
	return &MappedDS{
		dsMap: dsMap,
	}
}

func (m *MappedDS) Add(srcDS, targetDS string) *MappedDS {
	m.dsMap[srcDS] = targetDS
	return m
}

func (m *MappedDS) TargetDatastores(sourceVM *vcenter.VM) (map[string]string, error) {
	mappedDS := map[string]string{}
	for _, vmDisk := range sourceVM.Disks {
		targetDS, ok := m.dsMap[vmDisk.Datastore]
		if !ok {
			return nil, fmt.Errorf("could not find a target datastore for VM %s with source datastore %s: "+
				"ensure you add a corresponding datastore mapping to the config file",
				sourceVM.Name, vmDisk.Datastore)
		}
		mappedDS[vmDisk.Datastore] = targetDS
	}
	return mappedDS, nil
}
