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

const DefaultWorkerCount = 3

//counterfeiter:generate . BoshClient
type BoshClient interface {
	VMsAndStemcells(context.Context) ([]bosh.VM, error)
}

// NullBoshClient is a null object pattern when no bosh client is specified in the config
type NullBoshClient struct{}

// VMsAndStemcells returns an empty list
func (c NullBoshClient) VMsAndStemcells(context.Context) ([]bosh.VM, error) {
	return []bosh.VM{}, nil
}

type VM struct {
	Name string
	AZ   string

	// list of clusters within the source AZ that may contain the VM
	Clusters []string
}

// migrationResult holds the individual VM migration result
type migrationResult struct {
	id     int
	vmName string
	err    error
}

// Success the migration result is considered successful
func (mr migrationResult) Success() bool {
	return mr.err == nil
}

// FoundationMigrator orchestrates the entire migration of a foundation
type FoundationMigrator struct {
	WorkerCount     int
	AdditionalVMs   []bosh.VM
	updatableStdout *log.UpdatableStdout

	clientPool       *vcenter.Pool
	vmMigrator       *VMMigrator
	boshClient       BoshClient
	srcAZsToClusters map[string][]string
}

// NewFoundationMigrator creates a new initialized FoundationMigrator using the provided instances
func NewFoundationMigrator(
	clientPool *vcenter.Pool,
	boshClient BoshClient,
	vmMigrator *VMMigrator,
	srcAZsToClusters map[string][]string,
	out *log.UpdatableStdout) *FoundationMigrator {

	return &FoundationMigrator{
		WorkerCount:      DefaultWorkerCount,
		AdditionalVMs:    make([]bosh.VM, 0),
		clientPool:       clientPool,
		boshClient:       boshClient,
		vmMigrator:       vmMigrator,
		srcAZsToClusters: srcAZsToClusters,
		updatableStdout:  out,
	}
}

// NewFoundationMigratorFromConfig creates a new FoundationMigrator instance from the specified config
func NewFoundationMigratorFromConfig(c config.Config) (*FoundationMigrator, error) {
	l := log.WithoutContext()
	l.Debug("Creating vCenter client pool")
	clientPool := ConfigToVCenterClientPool(c)

	l.Debug("Creating BOSH client")
	boshClient := ConfigToBoshClient(c)

	l.Debug("Creating AZ cluster mappings")
	computeMap, err := ConfigToAZMapping(c)
	if err != nil {
		return nil, err
	}

	l.Debug("Creating source VM target spec converter")
	sourceVMConverter := converter.New(
		converter.NewMappedNetwork(c.NetworkMap),
		converter.NewMappedDatastore(c.DatastoreMap),
		converter.NewMappedCompute(computeMap))

	l.Debug("Creating VM migrator")
	hpConfig := ConfigToTargetHostPoolConfig(c)
	srcAZsToClusters := ConfigToSourceClustersByAZ(c)
	out := log.NewUpdatableStdout()
	destinationHostPool := vcenter.NewHostPool(clientPool, hpConfig)

	vmRelocator := vcenter.NewVMRelocator(clientPool, destinationHostPool, out).WithDryRun(c.DryRun)
	vmMigrator := NewVMMigrator(clientPool, sourceVMConverter, vmRelocator, out)

	l.Debug("Creating foundation migrator")
	fm := NewFoundationMigrator(clientPool, boshClient, vmMigrator, srcAZsToClusters, out)
	fm.AdditionalVMs = ConfigToAdditionalVMs(c)
	fm.WorkerCount = c.WorkerPoolSize
	return fm, nil
}

// Migrate executes the entire migration for all VMs
func (f *FoundationMigrator) Migrate(ctx context.Context) error {
	start := time.Now()
	l := log.WithoutContext()
	l.Infof("Starting foundation migration at %s", start.Format(time.RFC1123Z))

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
			l.Debugf("%s failed to migrate: %s", res.vmName, res.err)
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

func (f *FoundationMigrator) vmsToMigrate(ctx context.Context) ([]VM, error) {
	boshVMs, err := f.boshClient.VMsAndStemcells(ctx)
	if err != nil {
		return nil, err
	}
	boshVMs = append(boshVMs, f.AdditionalVMs...)

	var vms []VM
	for _, bvm := range boshVMs {
		vms = append(vms, VM{
			Name:     bvm.Name,
			AZ:       bvm.AZ,
			Clusters: f.srcAZsToClusters[bvm.AZ],
		})
	}
	return vms, nil
}

// ConfigToAZMapping creates the expanded source -> target AZ mappings used by the compute mapper
func ConfigToAZMapping(c config.Config) ([]converter.AZMapping, error) {
	var computeMap []converter.AZMapping
	for _, saz := range c.Compute.Source {
		for _, scl := range saz.Clusters {
			sm := converter.AZ{
				Name:         saz.Name,
				Datacenter:   saz.VCenter.Datacenter,
				Cluster:      scl.Name,
				ResourcePool: scl.ResourcePool,
			}
			taz := c.Compute.TargetByAZ(saz.Name)
			if taz == nil {
				return nil, fmt.Errorf("could not find a corresponding compute saz target named %s", saz.Name)
			}
			for _, tcl := range taz.Clusters {
				tm := converter.AZ{
					Name:         taz.Name,
					Datacenter:   taz.VCenter.Datacenter,
					Cluster:      tcl.Name,
					ResourcePool: tcl.ResourcePool,
				}
				computeMap = append(computeMap, converter.AZMapping{
					Source: sm,
					Target: tm,
				})
			}
		}
	}
	return computeMap, nil
}

// ConfigToTargetHostPoolConfig creates the required configuration format to create a target host pool
func ConfigToTargetHostPoolConfig(c config.Config) *vcenter.HostPoolConfig {
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
	return hpConfig
}

// ConfigToBoshClient creates an optional bosh client if configured, other a null object (pattern)
func ConfigToBoshClient(c config.Config) BoshClient {
	// if there's a configured optional BOSH config section then create a client
	if c.Bosh != nil {
		return bosh.New(c.Bosh.Host, c.Bosh.ClientID, c.Bosh.ClientSecret)
	}
	return NullBoshClient{}
}

// ConfigToVCenterClientPool creates a pool of VCenter clients for each AZ source and target
func ConfigToVCenterClientPool(c config.Config) *vcenter.Pool {
	clientPool := vcenter.NewPool()
	for _, az := range c.Compute.Source {
		clientPool.AddSource(az.Name, az.VCenter.Host, az.VCenter.Username, az.VCenter.Password, az.VCenter.Datacenter, az.VCenter.Insecure)
	}
	for _, az := range c.Compute.Target {
		clientPool.AddTarget(az.Name, az.VCenter.Host, az.VCenter.Username, az.VCenter.Password, az.VCenter.Datacenter, az.VCenter.Insecure)
	}
	return clientPool
}

// ConfigToAdditionalVMs creates a list of additional non-bosh VMs to migrate
func ConfigToAdditionalVMs(c config.Config) []bosh.VM {
	var additionalVMs []bosh.VM
	for az, vms := range c.AdditionalVMs {
		for _, v := range vms {
			additionalVMs = append(additionalVMs, bosh.VM{
				Name: v,
				AZ:   az,
			})
		}
	}
	return additionalVMs
}

// ConfigToSourceClustersByAZ creates a map of AZs to their source clusters
func ConfigToSourceClustersByAZ(c config.Config) map[string][]string {
	azToClusters := map[string][]string{}
	for _, az := range c.Compute.Source {
		var cls []string
		for _, cl := range az.Clusters {
			cls = append(cls, cl.Name)
		}
		azToClusters[az.Name] = cls
	}
	return azToClusters
}
