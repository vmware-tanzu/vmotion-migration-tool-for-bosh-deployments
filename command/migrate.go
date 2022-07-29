package command

import (
	"context"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/config"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate"
)

type Migrate struct {
	BoshClientSecret      string `long:"bosh-client-secret" env:"BOSH_CLIENT_SECRET" description:"BoshClient client secret" required:"true"`
	VCenterSourcePassword string `long:"source-password" env:"SOURCE_PASSWORD"  description:"password for source vcenter" required:"true"`
	VCenterTargetPassword string `long:"target-password" env:"TARGET_PASSWORD"  description:"password for target vcenter" required:"true"`
	ConfigFilePath        string `long:"config"  description:"path to the migrate.yml, defaults to ./migrate.yml"`
	DryRun                bool   `long:"dry-run"  description:"does not perform any migration operations when true"`
	Debug                 bool   `long:"debug"  description:"sets log level to debug"`
}

// Execute - runs the migration
func (m *Migrate) Execute([]string) error {
	log.Initialize(m.Debug)
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

	c.Source.Bosh.ClientSecret = m.BoshClientSecret
	c.Source.VCenter.Password = m.VCenterSourcePassword
	c.Target.VCenter.Password = m.VCenterTargetPassword
	c.DryRun = m.DryRun

	log.WithoutContext().Debugf("Combined config: \n%s", c)
	return c, nil
}
