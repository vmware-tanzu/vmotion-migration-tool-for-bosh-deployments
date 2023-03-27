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
	ConfigFilePath string `long:"config"  description:"path to the migrate.yml, defaults to ./migrate.yml"`
	DryRun         bool   `long:"dry-run"  description:"does not perform any migration operations when true"`
	Debug          bool   `long:"debug"  description:"sets log level to debug"`
	RedactSecrets  bool   `long:"no-redact" description:"do not redact sensitive information when printing debug logs"`
}

// Execute - runs the migration
func (m *Migrate) Execute([]string) error {
	log.Initialize(m.Debug, m.RedactSecrets)
	ctx := context.Background()

	c, err := m.combinedConfig()
	if err != nil {
		return err
	}

	fm, err := migrate.NewFoundationMigrator(ctx, c)
	if err != nil {
		return err
	}
	return fm.Migrate(ctx)
}

func (m *Migrate) combinedConfig() (config.Config, error) {
	if m.ConfigFilePath == "" {
		m.ConfigFilePath = "migrate.yml"
	}
	c, err := config.NewConfigFromFile(m.ConfigFilePath)
	if err != nil {
		return config.Config{}, err
	}
	log.WithoutContext().Debugf("Combined config: \n%s", c)
	return c, nil
}
