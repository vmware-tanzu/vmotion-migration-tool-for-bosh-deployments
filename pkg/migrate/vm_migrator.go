/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package migrate

import (
	"context"
	"errors"
	"fmt"

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
	m.printProcessing(ctx, srcVMName, "preparing")
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
				m.printSuccess(ctx, srcVMName, "already migrated, skipping")
			} else {
				m.printSuccess(ctx, srcVMName, "not found in source vCenter, skipping")
			}
			return nil
		}
		m.printFailure(ctx, srcVMName, err)
		return err
	}

	vmTargetSpec, err := m.sourceVMConverter.TargetSpec(srcVM)
	if err != nil {
		m.printFailure(ctx, srcVMName, err)
		return err
	}

	err = m.vmRelocator.RelocateVM(ctx, srcVM, vmTargetSpec)
	if err != nil {
		m.printFailure(ctx, srcVMName, err)
		return err
	}

	m.printSuccess(ctx, srcVMName, "done")
	return nil
}

const greenCheck = "✅"
const redX = "❌"

func (m *VMMigrator) printFailure(ctx context.Context, srcVMName string, err error) {
	log.FromContext(ctx).Errorf("%s failed: %s", srcVMName, err)
	m.updatableStdout.PrintUpdatablef(srcVMName, "%s %s - %s", srcVMName, redX, err)
}

func (m *VMMigrator) printProcessing(ctx context.Context, srcVMName, msg string) {
	log.FromContext(ctx).Infof("%s processing: %s", srcVMName, msg)
	m.updatableStdout.PrintUpdatablef(srcVMName, "%s - %s", srcVMName, fmt.Sprintf("%-40s", msg))
}

func (m *VMMigrator) printSuccess(ctx context.Context, srcVMName, msg string) {
	log.FromContext(ctx).Infof("%s done: %s", srcVMName, msg)
	m.updatableStdout.PrintUpdatablef(srcVMName, "%s %s - %-40s", srcVMName, greenCheck, msg)
}
