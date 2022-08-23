/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"os"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"gopkg.in/yaml.v3"
)

func NewConfigFromFile(configFilePath string) (Config, error) {
	log.WithoutContext().Debugf("Reading config file: %s", configFilePath)
	buf, err := os.ReadFile(configFilePath)
	if err != nil {
		return Config{}, err
	}

	c := Config{}
	err = yaml.Unmarshal(buf, &c)
	if err != nil {
		return Config{}, fmt.Errorf("in file %q: %w", configFilePath, err)
	}

	// if target not specified, assume source and destination vcenter are same
	if c.Target.VCenter.Host == "" {
		c.Target.VCenter.Host = c.Source.VCenter.Host
		c.Target.VCenter.Username = c.Source.VCenter.Username
		c.Target.VCenter.Password = c.Source.VCenter.Password
		c.Target.VCenter.Insecure = c.Source.VCenter.Insecure
	}

	return c, nil
}

type VCenter struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string
	Insecure bool `yaml:"insecure"`
}

type Bosh struct {
	Host         string `yaml:"host"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string
}

type Target struct {
	VCenter
	Datacenter string `yaml:"datacenter"`
	Cluster    string `yaml:"cluster"`   // assumes single cluster foundation
	Datastore  string `yaml:"datastore"` // assumes single cluster foundation
}

type Source struct {
	VCenter
	Bosh
	Datacenter string `yaml:"datacenter"`
}

type Config struct {
	Source          Source
	Target          Target
	DryRun          bool
	ResourcePoolMap map[string]string `yaml:"resource_pools"`
	NetworkMap      map[string]string `yaml:"networks"`
}

// String used primarily for debug logging
func (c Config) String() string {
	// easier to compare as yaml
	b, err := yaml.Marshal(c)
	if err != nil {
		return err.Error()
	}
	return string(b)
}
