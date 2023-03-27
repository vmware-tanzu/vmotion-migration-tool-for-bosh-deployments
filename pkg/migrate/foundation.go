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
	VMsAndStemcells(context.Context) ([]bosh.VM, error)
}

type NullBoshClient struct{}

func (c NullBoshClient) VMsAndStemcells(context.Context) ([]bosh.VM, error) {
	return []bosh.VM{}, nil
}

type migrationResult struct {
	id     int
	vmName string
	err    error
}

func (mr migrationResult) Success() bool {
	return mr.err == nil
}

type FoundationMigrator struct {
	WorkerCount int

	additionalVMs   []bosh.VM
	clientPool      *vcenter.Pool
	vmMigrator      *VMMigrator
	mappedCompute   *converter.MappedCompute
	boshClient      BoshClient
	updatableStdout *log.UpdatableStdout
}

func NewFoundationMigrator(ctx context.Context, c config.Config) (*FoundationMigrator, error) {
	l := log.WithoutContext()
	l.Infof("Preparing foundation migration at %s", time.Now().Format(time.RFC1123Z))

	// create a pool of VCenter clients for each AZ source and target
	l.Debug("Creating vCenter client pool")
	clientPool := vcenter.NewPool()
	for _, az := range c.Compute.Source {
		clientPool.AddSource(az.Name, az.VCenter.Host, az.VCenter.Username, az.VCenter.Password, az.VCenter.Datacenter, az.VCenter.Insecure)
	}
	for _, az := range c.Compute.Target {
		clientPool.AddTarget(az.Name, az.VCenter.Host, az.VCenter.Username, az.VCenter.Password, az.VCenter.Datacenter, az.VCenter.Insecure)
	}

	// if there's a configured optional BOSH config section then create a client
	l.Debug("Creating BOSH client")
	var boshClient BoshClient
	if c.Bosh != nil {
		boshClient = bosh.New(c.Bosh.Host, c.Bosh.ClientID, c.Bosh.ClientSecret)
	} else {
		boshClient = NullBoshClient{}
	}

	// generate a unique mapping for each az src/target cluster combo
	l.Debug("Creating AZ cluster mappings")
	var computeMap []converter.AZMapping
	for _, az := range c.Compute.Source {
		for _, cc := range az.Clusters {
			sm := converter.AZ{
				Name:         az.Name,
				Datacenter:   az.VCenter.Datacenter,
				Cluster:      cc.Name,
				ResourcePool: cc.ResourcePool,
			}
			taz := c.Compute.TargetByAZ(az.Name)
			if taz == nil {
				return nil, fmt.Errorf("could not find a corresponding compute az target named %s", az.Name)
			}
			for _, tcc := range taz.Clusters {
				tm := converter.AZ{
					Name:         taz.Name,
					Datacenter:   taz.VCenter.Datacenter,
					Cluster:      tcc.Name,
					ResourcePool: tcc.ResourcePool,
				}
				m := converter.AZMapping{
					Source: sm,
					Target: tm,
				}
				computeMap = append(computeMap, m)
			}
		}
	}

	// create a VM converter instance
	mappedCompute := converter.NewMappedCompute(computeMap)
	sourceVMConverter := converter.New(
		converter.NewMappedNetwork(c.NetworkMap),
		converter.NewMappedDatastore(c.DatastoreMap),
		mappedCompute)

	// create a host pool for each target AZ/cluster
	l.Debug("Creating vCenter host pools")
	hpConfig := &vcenter.HostPoolConfig{}
	hpConfig.AZs = make(map[string]vcenter.HostPoolAZ, len(c.Compute.Target))
	for _, t := range c.Compute.Target {
		var cls []string
		for _, a := range t.Clusters {
			cls = append(cls, a.Name)
		}
		hpConfig.AZs[t.Name] = vcenter.HostPoolAZ{
			Clusters: cls,
		}
	}

	destinationHostPool := vcenter.NewHostPool(clientPool, hpConfig)
	err := destinationHostPool.Initialize(ctx)
	if err != nil {
		return nil, err
	}

	// add additional VMs from config
	var additionalVMs []bosh.VM
	for az, vms := range c.AdditionalVMs {
		for _, v := range vms {
			additionalVMs = append(additionalVMs, bosh.VM{
				Name: v,
				AZ:   az,
			})
		}
	}

	l.Debug("Creating foundation migrator")
	out := log.NewUpdatableStdout()
	vmRelocator := vcenter.NewVMRelocator(clientPool, destinationHostPool, out).WithDryRun(c.DryRun)
	vmMigrator := NewVMMigrator(clientPool, sourceVMConverter, vmRelocator, out)

	fm := &FoundationMigrator{
		clientPool:    clientPool,
		boshClient:    boshClient,
		vmMigrator:    vmMigrator,
		mappedCompute: mappedCompute,
		additionalVMs: additionalVMs,
		WorkerCount:   c.WorkerPoolSize,
	}
	return fm, nil
}

func (f *FoundationMigrator) Migrate(ctx context.Context) error {
	start := time.Now()
	log.WithoutContext().Infof("Starting foundation migration at %s", start.Format(time.RFC1123Z))

	defer f.clientPool.Close(ctx)

	vms, err := f.vmsToMigrate(ctx)
	if err != nil {
		return err
	}

	vmCount := len(vms)
	results := make(chan migrationResult, vmCount)

	workers := worker.NewPool(f.WorkerCount)
	workers.Start(ctx)

	for i, vm := range vms {
		i := i + 1 // closure and make it 1 based
		v := vm    // closure
		workers.AddTask(func(taskCtx context.Context) {
			err := f.vmMigrator.Migrate(taskCtx, v)
			results <- migrationResult{
				id:     i,
				vmName: v.Name,
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

func (f *FoundationMigrator) vmsToMigrate(ctx context.Context) ([]bosh.VM, error) {
	vms, err := f.boshClient.VMsAndStemcells(ctx)
	if err != nil {
		return nil, err
	}
	vms = append(vms, f.additionalVMs...)
	return vms, nil
}
