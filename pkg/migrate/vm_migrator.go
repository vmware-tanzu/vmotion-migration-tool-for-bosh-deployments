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
	FindVM(ctx context.Context, datacenter, cluster, vmName string) (*vcenter.VM, error)
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

	srcVM, err := m.sourceVCenter.FindVM(ctx, srcDatacenter, "", srcVMName)
	if err != nil {
		var e *vcenter.VMNotFoundError
		if errors.As(err, &e) {
			m.updatableStdout.PrintUpdatablef(srcVMName, "%s - not found in source vCenter, skipping", srcVMName)
			// assume it's already been previously migrated (handle missing VMs via BOSH)
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

	return m.vmRelocator.RelocateVM(ctx, srcVM, vmTargetSpec)
}
