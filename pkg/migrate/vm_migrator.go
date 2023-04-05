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
	Datacenter() string
	FindVMInClusters(ctx context.Context, az, vmNameOrPath string, clusters []string) (*vcenter.VM, error)
}

type UpdatableLogger interface {
	PrintUpdatablef(id, format string, a ...interface{})
}

type VMMigrator struct {
	sourceVMConverter *converter.Converter
	clientPool        *vcenter.Pool
	vmRelocator       VMRelocator
	updatableStdout   UpdatableLogger
}

func NewVMMigrator(clientPool *vcenter.Pool, sourceVMConverter *converter.Converter, vmRelocator VMRelocator, updatableStdout UpdatableLogger) *VMMigrator {
	return &VMMigrator{
		clientPool:        clientPool,
		sourceVMConverter: sourceVMConverter,
		vmRelocator:       vmRelocator,
		updatableStdout:   updatableStdout,
	}
}

func (m *VMMigrator) Migrate(ctx context.Context, sourceVM VM) error {
	sourceClient := m.clientPool.GetSourceClientByAZ(sourceVM.AZ)
	if sourceClient == nil {
		return fmt.Errorf("could not find source vcenter client for VM %s in AZ %s", sourceVM.Name, sourceVM.AZ)
	}
	return m.MigrateVMToTarget(ctx, sourceClient, sourceVM)
}

func (m *VMMigrator) MigrateVMToTarget(ctx context.Context, sourceClient VCenterClient, sourceVM VM) error {
	m.printProcessing(ctx, sourceVM.Name, "preparing")

	// find the VM to migrate but only look in the source cluster(s) as it may have already been moved
	v, err := sourceClient.FindVMInClusters(ctx, sourceVM.AZ, sourceVM.Name, sourceVM.Clusters)
	if err != nil {
		var e *vcenter.VMNotFoundError
		if errors.As(err, &e) {
			m.printSuccess(ctx, sourceVM.Name, "not found in source vCenter, skipping")
			return nil
		}
		m.printFailure(ctx, sourceVM.Name, err)
		return err
	}

	vmTargetSpec, err := m.sourceVMConverter.TargetSpec(v)
	if err != nil {
		m.printFailure(ctx, sourceVM.Name, err)
		return err
	}

	err = m.vmRelocator.RelocateVM(ctx, v, vmTargetSpec)
	if err != nil {
		m.printFailure(ctx, sourceVM.Name, err)
		return err
	}

	m.printSuccess(ctx, sourceVM.Name, "done")
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
