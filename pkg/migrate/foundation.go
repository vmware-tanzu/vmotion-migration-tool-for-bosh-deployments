/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package migrate

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"
	"time"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/bosh"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/config"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/duration"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate/converter"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/worker"
)

//counterfeiter:generate . BoshClient
type BoshClient interface {
	VMsAndStemcells(context.Context) ([]string, error)
}

type NullBoshClient struct{}

func (c NullBoshClient) VMsAndStemcells(context.Context) ([]string, error) {
	return []string{}, nil
}

type FoundationMigrator struct {
	WorkerCount   int
	AdditionalVMs []string

	sourceDatacenter string
	sourceCluster    string
	vmMigrator       *VMMigrator
	boshClient       BoshClient
	updatableStdout  *log.UpdatableStdout
}

type migrationResult struct {
	id     int
	vmName string
	err    error
}

func (mr migrationResult) Success() bool {
	return mr.err == nil
}

func NewFoundationMigrator(
	srcDatacenter string, boshClient BoshClient, vmMigrator *VMMigrator, updatableStdout *log.UpdatableStdout) *FoundationMigrator {
	return &FoundationMigrator{
		sourceDatacenter: srcDatacenter,
		sourceCluster:    "",
		WorkerCount:      5,
		boshClient:       boshClient,
		vmMigrator:       vmMigrator,
		updatableStdout:  updatableStdout,
	}
}

func RunFoundationMigrationWithConfig(c config.Config, ctx context.Context) error {
	destinationVCenter := vcenter.New(
		c.Target.VCenter.Host, c.Target.VCenter.Username, c.Target.VCenter.Password, c.Target.VCenter.Insecure)
	defer destinationVCenter.Logout(ctx)

	sourceVCenter := vcenter.New(
		c.Source.VCenter.Host, c.Source.VCenter.Username, c.Source.VCenter.Password, c.Source.VCenter.Insecure)
	defer sourceVCenter.Logout(ctx)

	// if there's a configured optional BOSH config section then create a client
	var boshClient BoshClient
	if c.Bosh != nil {
		boshClient = bosh.New(c.Bosh.Host, c.Bosh.ClientID, c.Bosh.ClientSecret)
	} else {
		boshClient = NullBoshClient{}
	}

	sourceVMConverter := converter.New(
		converter.NewMappedNetwork(c.NetworkMap),
		converter.NewMappedResourcePool(c.ResourcePoolMap),
		converter.NewMappedDatastore(c.DatastoreMap),
		converter.NewMappedCluster(c.ClusterMap),
		c.Target.Datacenter)

	destinationHostPool := vcenter.NewHostPool(destinationVCenter, c.Target.Datacenter)
	err := destinationHostPool.Initialize(ctx)
	if err != nil {
		return err
	}

	out := log.NewUpdatableStdout()
	vmRelocator := vcenter.NewVMRelocator(sourceVCenter, destinationVCenter, destinationHostPool, out).WithDryRun(c.DryRun)
	vmMigrator := NewVMMigrator(sourceVCenter, destinationVCenter, sourceVMConverter, vmRelocator, out)

	fm := NewFoundationMigrator(c.Source.Datacenter, boshClient, vmMigrator, out)
	fm.WorkerCount = c.WorkerPoolSize
	fm.AdditionalVMs = c.AdditionalVMs

	return fm.Migrate(ctx)
}

func (f *FoundationMigrator) Migrate(ctx context.Context) error {
	log.WithoutContext().Infof("Migrating all bosh managed VMs from %s to %s",
		f.vmMigrator.sourceVCenter.HostName(), f.vmMigrator.targetVCenter.HostName())

	start := time.Now()

	vms, err := f.vmsToMigrate(ctx)
	if err != nil {
		return err
	}

	vmCount := len(vms)
	results := make(chan migrationResult, vmCount)

	workers := worker.NewPool(f.WorkerCount)
	workers.Start(ctx)

	for i, vm := range vms {
		i := i + 1   // closure and make it 1 based
		vmName := vm // closure
		workers.AddTask(func(taskCtx context.Context) {
			err := f.vmMigrator.Migrate(taskCtx, f.sourceDatacenter, vmName)
			results <- migrationResult{
				id:     i,
				vmName: vmName,
				err:    err,
			}
		})
	}

	failCount := 0
	for i := 0; i < vmCount; i++ {
		res := <-results
		if !res.Success() {
			failCount++
		}
	}
	close(results)

	f.updatableStdout.Println()
	f.updatableStdout.Printf("Migrated %d out of %d VMs", vmCount-failCount, vmCount)
	f.updatableStdout.Printf("Total runtime: %s", duration.HumanReadable(time.Since(start)))

	if failCount > 0 {
		return fmt.Errorf("failed to migrate %d VMs, see run output for more details", failCount)
	}

	return nil
}

func (f *FoundationMigrator) vmsToMigrate(ctx context.Context) ([]string, error) {
	vms, err := f.boshClient.VMsAndStemcells(ctx)
	if err != nil {
		return nil, err
	}
	vms = append(vms, f.AdditionalVMs...)
	return vms, nil
}
