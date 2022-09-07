/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package command

import (
	"context"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate"
)

type Revert struct {
	Migrate
}

// Execute - runs the migration in reverse
func (r *Revert) Execute([]string) error {
	log.Initialize(r.Debug, r.RedactSecrets)
	ctx := context.Background()

	c, err := r.combinedConfig()
	if err != nil {
		return err
	}

	return migrate.RunFoundationMigrationWithConfig(c.Reversed(), ctx)
}
