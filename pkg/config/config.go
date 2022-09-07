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

	if c.WorkerPoolSize == 0 {
		c.WorkerPoolSize = 3
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
}

type Source struct {
	VCenter
	Datacenter string `yaml:"datacenter"`
}

type Config struct {
	Source          Source
	Target          Target
	Bosh            *Bosh `yaml:"bosh"`
	DryRun          bool
	WorkerPoolSize  int               `yaml:"worker_pool_size"`
	ResourcePoolMap map[string]string `yaml:"resource_pools"`
	NetworkMap      map[string]string `yaml:"networks"`
	DatastoreMap    map[string]string `yaml:"datastores"`
	ClusterMap      map[string]string `yaml:"clusters"`
	AdditionalVMs   []string          `yaml:"additional_vms"`
}

func (c Config) Reversed() Config {
	rc := Config{
		Source: Source{
			VCenter: VCenter{
				Host:     c.Target.Host,
				Username: c.Target.Username,
				Password: c.Target.Password,
				Insecure: c.Target.Insecure,
			},
			Datacenter: c.Target.Datacenter,
		},
		Target: Target{
			VCenter: VCenter{
				Host:     c.Source.Host,
				Username: c.Source.Username,
				Password: c.Source.Password,
				Insecure: c.Source.Insecure,
			},
			Datacenter: c.Source.Datacenter,
		},
		DryRun:         c.DryRun,
		WorkerPoolSize: c.WorkerPoolSize,
		AdditionalVMs:  c.AdditionalVMs,
	}

	rc.ResourcePoolMap = make(map[string]string, len(c.ResourcePoolMap))
	for k, v := range c.ResourcePoolMap {
		rc.ResourcePoolMap[v] = k
	}

	rc.NetworkMap = make(map[string]string, len(c.NetworkMap))
	for k, v := range c.NetworkMap {
		rc.NetworkMap[v] = k
	}

	rc.DatastoreMap = make(map[string]string, len(c.DatastoreMap))
	for k, v := range c.DatastoreMap {
		rc.DatastoreMap[v] = k
	}

	rc.ClusterMap = make(map[string]string, len(c.ClusterMap))
	for k, v := range c.ClusterMap {
		rc.ClusterMap[v] = k
	}

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
