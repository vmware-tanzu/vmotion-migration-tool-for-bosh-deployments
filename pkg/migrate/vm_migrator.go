/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package migrate

import (
	"context"
	"errors"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate/converter"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

//counterfeiter:generate . VMRelocator
type VMRelocator interface {
	RelocateVM(ctx context.Context, srcVM *vcenter.VM, vmTargetSpec *vcenter.TargetSpec) error
}

//counterfeiter:generate . VCenterClient
type VCenterClient interface {
	HostName() string
	FindVM(ctx context.Context, datacenter, vmName string) (*vcenter.VM, error)
}

type VMMigrator struct {
	sourceVMConverter *converter.Converter
	sourceVCenter     VCenterClient
	targetVCenter     VCenterClient
	vmRelocator       VMRelocator
	updatableStdout   *log.UpdatableStdout
}

func NewVMMigrator(sourceVCenter, targetVCenter VCenterClient, sourceVMConverter *converter.Converter, vmRelocator VMRelocator, updatableStdout *log.UpdatableStdout) *VMMigrator {
	return &VMMigrator{
		sourceVMConverter: sourceVMConverter,
		sourceVCenter:     sourceVCenter,
		targetVCenter:     targetVCenter,
		vmRelocator:       vmRelocator,
		updatableStdout:   updatableStdout,
	}
}

func (m *VMMigrator) Migrate(ctx context.Context, srcDatacenter, srcVMName string) error {
	l := log.FromContext(ctx)
	l.Infof("Migrating VM %s from %s to %s",
		srcVMName, m.sourceVCenter.HostName(), m.targetVCenter.HostName())

	// find the VM to migrate in the source
	srcVM, err := m.sourceVCenter.FindVM(ctx, srcDatacenter, srcVMName)
	if err != nil {
		var e *vcenter.VMNotFoundError
		if errors.As(err, &e) {
			destVM, _ := m.targetVCenter.FindVM(ctx, m.sourceVMConverter.TargetDatacenter(), srcVMName)
			if destVM != nil {
				m.updatableStdout.PrintUpdatablef(srcVMName, "%s - already migrated, skipping", srcVMName)
			} else {
				m.updatableStdout.PrintUpdatablef(srcVMName, "%s - not found in source vCenter, skipping", srcVMName)
			}
			return nil
		}
		return err
	}

	vmTargetSpec, err := m.sourceVMConverter.TargetSpec(srcVM)
	if err != nil {
		return err
	}

	l.Debugf("Source VM:\n%+v", srcVM)
	l.Debugf("Target VM:\n%+v", vmTargetSpec)

	err = m.vmRelocator.RelocateVM(ctx, srcVM, vmTargetSpec)
	if err != nil {
		return err
	}

	m.updatableStdout.PrintUpdatablef(srcVMName, "%s - done", srcVMName)
	return nil
}
