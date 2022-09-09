/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package command

import (
	"context"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/config"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate"
)

type Migrate struct {
	VCenterSourcePassword string `long:"source-password" env:"SOURCE_PASSWORD"  description:"password for source vcenter" required:"true"`
	VCenterTargetPassword string `long:"target-password" env:"TARGET_PASSWORD"  description:"password for target vcenter"`
	BoshClientSecret      string `long:"bosh-client-secret" env:"BOSH_CLIENT_SECRET" description:"BoshClient client secret"`
	ConfigFilePath        string `long:"config"  description:"path to the migrate.yml, defaults to ./migrate.yml"`
	DryRun                bool   `long:"dry-run"  description:"does not perform any migration operations when true"`
	Debug                 bool   `long:"debug"  description:"sets log level to debug"`
	RedactSecrets         bool   `long:"no-redact" description:"do not redact sensitive information when printing debug logs"`
}

// Execute - runs the migration
func (m *Migrate) Execute([]string) error {
	log.Initialize(m.Debug, m.RedactSecrets)
	ctx := context.Background()

	c, err := m.combinedConfig()
	if err != nil {
		return err
	}

	return migrate.RunFoundationMigrationWithConfig(c, ctx)
}

func (m *Migrate) combinedConfig() (config.Config, error) {
	if m.ConfigFilePath == "" {
		m.ConfigFilePath = "migrate.yml"
	}
	c, err := config.NewConfigFromFile(m.ConfigFilePath)
	if err != nil {
		return config.Config{}, err
	}
	m.overwriteFromCommandLine(&c)
	log.WithoutContext().Debugf("Combined config: \n%s", c)
	return c, nil
}

func (m *Migrate) overwriteFromCommandLine(c *config.Config) {
	if c.Bosh != nil {
		c.Bosh.ClientSecret = m.BoshClientSecret
	}
	c.Source.VCenter.Password = m.VCenterSourcePassword
	c.DryRun = m.DryRun

	// in case source and target vcenter are the same
	if m.VCenterTargetPassword == "" {
		c.Target.VCenter.Password = m.VCenterSourcePassword
	} else {
		c.Target.VCenter.Password = m.VCenterTargetPassword
	}
}
