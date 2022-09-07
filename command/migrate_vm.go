/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package command

import (
	"context"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate/converter"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

type MigrateVM struct {
	SourceHost       string `long:"source-vcenter-host" env:"SOURCE_VCENTER_HOST"  description:"source vcenter hostname" required:"true"`
	SourceDatacenter string `long:"source-datacenter" description:"source datacenter name (where the vm is now)" required:"true"`
	SourceUsername   string `long:"source-username" env:"SOURCE_USERNAME"  description:"username for source vcenter" required:"true"`
	SourcePassword   string `long:"source-password" env:"SOURCE_PASSWORD"  description:"password for source vcenter" required:"true"`
	SourceVM         string `long:"source-vmname" description:"source virtual machine name" required:"true"`
	SourceInsecure   bool   `long:"source-insecure" env:"SOURCE_INSECURE" description:"true if using a self-signed cert" required:"false"`

	TargetHost         string `long:"target-vcenter-host" env:"TARGET_VCENTER_HOST"  description:"target vcenter hostname" required:"true"`
	TargetDatacenter   string `long:"target-datacenter" description:"target datacenter name (where you want the vm to go)" required:"true"`
	TargetResourcePool string `long:"target-resourcepool" description:"target resource pool name"`
	TargetUsername     string `long:"target-username" env:"TARGET_USERNAME"  description:"username for target vcenter" required:"true"`
	TargetPassword     string `long:"target-password" env:"TARGET_PASSWORD"  description:"password for target vcenter" required:"true"`
	TargetInsecure     bool   `long:"target-insecure" env:"TARGET_INSECURE" description:"true if using a self-signed cert" required:"false"`

	NetworkMapping   map[string]string `long:"network-mapping" description:"source to target network name mappings"`
	DatastoreMapping map[string]string `long:"datastore-mapping" description:"source to target datastore name mappings" required:"true"`
	ClusterMapping   map[string]string `long:"cluster-mapping" description:"source to target cluster name mappings" required:"true"`

	DryRun        bool `long:"dry-run"  description:"does not perform any migration operations when true"`
	Debug         bool `long:"debug"  description:"sets log level to debug"`
	RedactSecrets bool `long:"no-redact" description:"do not redact sensitive information when printing debug logs"`
}

func (m *MigrateVM) Execute([]string) error {
	log.Initialize(m.Debug, m.RedactSecrets)
	ctx := context.Background()

	destinationVCenter := vcenter.New(m.TargetHost, m.TargetUsername, m.TargetPassword, m.TargetInsecure)
	defer destinationVCenter.Logout(ctx)

	sourceVCenter := vcenter.New(m.SourceHost, m.SourceUsername, m.SourcePassword, m.SourceInsecure)
	defer sourceVCenter.Logout(ctx)

	sourceVMConverter := converter.New(
		converter.NewMappedNetwork(m.NetworkMapping),
		converter.NewExplicitResourcePool(m.TargetResourcePool),
		converter.NewMappedDatastore(m.DatastoreMapping),
		converter.NewMappedCluster(m.ClusterMapping),
		m.TargetDatacenter)

	destinationHostPool := vcenter.NewHostPool(destinationVCenter, m.TargetDatacenter)
	err := destinationHostPool.Initialize(ctx)
	if err != nil {
		return err
	}

	out := log.NewUpdatableStdout()
	vmRelocator := vcenter.NewVMRelocator(sourceVCenter, destinationVCenter, destinationHostPool, out).WithDryRun(m.DryRun)

	migrator := migrate.NewVMMigrator(sourceVCenter, destinationVCenter, sourceVMConverter, vmRelocator, out)
	return migrator.Migrate(ctx, m.SourceDatacenter, m.SourceVM)
}
