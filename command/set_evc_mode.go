package command

import (
	"context"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/evc"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/config"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
)

type SetEVCMode struct {
	BoshClientSecret      string `long:"bosh-client-secret" env:"BOSH_CLIENT_SECRET" description:"BoshClient client secret" required:"true"`
	VCenterSourcePassword string `long:"source-password" env:"SOURCE_PASSWORD"  description:"password for source vcenter" required:"true"`
	EVCMode               string `long:"mode" description:"EVC mode to set: intel-merom, intel-sandybridge etc" required:"true"`
	ConfigFilePath        string `long:"config"  description:"path to the migrate.yml, defaults to ./migrate.yml"`
	DryRun                bool   `long:"dry-run"  description:"does not perform any migration operations when true"`
	Debug                 bool   `long:"debug"  description:"sets log level to debug"`
}

// Execute - sets the EVC mode on all VMs
func (e *SetEVCMode) Execute([]string) error {
	log.Initialize(e.Debug)
	ctx := context.Background()

	c, err := e.combinedConfig()
	if err != nil {
		return err
	}

	return evc.SetOnAllSourceVMs(e.EVCMode, c, ctx)
}

func (e *SetEVCMode) combinedConfig() (config.Config, error) {
	if e.ConfigFilePath == "" {
		e.ConfigFilePath = "migrate.yml"
	}
	c, err := config.NewConfigFromFile(e.ConfigFilePath)
	if err != nil {
		return config.Config{}, err
	}

	c.Source.Bosh.ClientSecret = e.BoshClientSecret
	c.Source.VCenter.Password = e.VCenterSourcePassword
	c.DryRun = e.DryRun

	log.WithoutContext().Debugf("Combined config: \n%s", c)
	return c, nil
}
