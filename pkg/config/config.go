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

	if c.WorkerPoolSize == 0 {
		c.WorkerPoolSize = 3
	}

	return c, nil
}

type Bosh struct {
	Host         string `yaml:"host"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string
}

type VCenter struct {
	Host       string `yaml:"host"`
	Username   string `yaml:"username"`
	Password   string
	Insecure   bool   `yaml:"insecure"`
	Datacenter string `yaml:"datacenter"`
}

type ComputeCluster struct {
	Name         string `yaml:"name"`
	ResourcePool string `yaml:"resource_pool"`
}

type ComputeAZ struct {
	Name     string           `yaml:"name"`
	VCenter  *VCenter         `yaml:"vcenter"`
	Clusters []ComputeCluster `yaml:"clusters"`
}

type Compute struct {
	Source []ComputeAZ `yaml:"source"`
	Target []ComputeAZ `yaml:"target"`
}

func (c Compute) TargetByAZ(azName string) *ComputeAZ {
	for _, taz := range c.Target {
		if taz.Name == azName {
			return &taz
		}
	}
	return nil
}

type Config struct {
	Bosh *Bosh `yaml:"bosh"`

	DryRun         bool
	WorkerPoolSize int      `yaml:"worker_pool_size"`
	AdditionalVMs  []string `yaml:"additional_vms"`

	NetworkMap   map[string]string `yaml:"networks"`
	DatastoreMap map[string]string `yaml:"datastores"`
	Compute      Compute           `yaml:"compute"`
}

func (c Config) Reversed() Config {
	rc := Config{
		DryRun:         c.DryRun,
		WorkerPoolSize: c.WorkerPoolSize,
		AdditionalVMs:  c.AdditionalVMs,
	}

	rc.NetworkMap = make(map[string]string, len(c.NetworkMap))
	for k, v := range c.NetworkMap {
		rc.NetworkMap[v] = k
	}

	rc.DatastoreMap = make(map[string]string, len(c.DatastoreMap))
	for k, v := range c.DatastoreMap {
		rc.DatastoreMap[v] = k
	}

	rc.Compute.Source = c.Compute.Target
	rc.Compute.Target = c.Compute.Source

	if c.Bosh != nil {
		rc.Bosh = &Bosh{
			Host:         c.Bosh.Host,
			ClientID:     c.Bosh.ClientID,
			ClientSecret: c.Bosh.ClientSecret,
		}
	}

	return rc
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
