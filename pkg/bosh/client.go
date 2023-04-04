/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package bosh

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"

	"github.com/cloudfoundry-community/gogobosh"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
)

//counterfeiter:generate . GogoBoshClient
type GogoBoshClient interface {
	GetDeploymentVMs(deployment string) ([]gogobosh.VM, error)
	GetCloudConfig(latest bool) ([]gogobosh.Cfg, error)
	GetDeployments() ([]gogobosh.Deployment, error)
	GetStemcells() ([]gogobosh.Stemcell, error)
}

type Client struct {
	ClientID     string
	ClientSecret string
	Environment  string
	client       GogoBoshClient
}

func New(environment, clientID, clientSecret string) *Client {
	return &Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Environment:  environment,
	}
}

func NewFromGogoBoshClient(client GogoBoshClient) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) VMsAndStemcells(ctx context.Context) ([]VM, error) {
	l := log.FromContext(ctx)

	client, err := c.getOrCreateUnderlyingClient()
	if err != nil {
		return nil, err
	}

	l.Debug("Getting BOSH cloud config")
	configs, err := client.GetCloudConfig(true)
	if err != nil {
		return nil, err
	}

	// find the default cloud config
	var cc string
	for _, cfg := range configs {
		if cfg.Name == "default" && cfg.Type == "cloud" {
			cc = cfg.Content
		}
	}
	if cc == "" {
		return nil, errors.New("could not find BOSH default cloud config")
	}

	cloudConfig := CloudConfig{}
	err = yaml.Unmarshal([]byte(cc), &cloudConfig)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal BOSH cloud config: %w", err)
	}

	// create map of CPI IDs -> AZs/Clusters
	cpiToAZ := map[string]string{}
	for _, az := range cloudConfig.AZs {
		// just add the first AZ - we don't care which AZ as long as it's maps to a vCenter/CPI
		if _, ok := cpiToAZ[az.Name]; !ok {
			cpiToAZ[az.CPI] = az.Name
		}
	}

	l.Debug("Getting all BOSH managed stemcells")
	stemcells, err := client.GetStemcells()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get bosh stemcells, this can happen because of incorrect login details: %w", err)
	}

	var result []VM
	for _, s := range stemcells {
		l.Debugf("  %s - %s", s.Name, s.CID)
		az, ok := cpiToAZ[s.CPI]
		if !ok {
			return nil, fmt.Errorf("could not find a CPI to AZ mapping for stemcell %s", s.CID)
		}
		v := VM{
			Name: s.CID,
			AZ:   az,
		}
		result = append(result, v)
	}

	deployments, err := client.GetDeployments()
	if err != nil {
		return nil, fmt.Errorf("failed to get bosh deployments: %w", err)
	}

	for _, d := range deployments {
		l.Infof("Found deployment %s", d.Name)
		vms, err := client.GetDeploymentVMs(d.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get deployment %s VMs: %w", d.Name, err)
		}
		l.Infof("With %d BOSH managed VMs", len(vms))

		for _, vm := range vms {
			instanceName := vm.JobName + "/" + vm.ID
			l.Debugf("  %s - %s", vm.VMCID, instanceName)
			v := VM{
				Name: vm.VMCID,
				AZ:   vm.AZ,
			}
			result = append(result, v)
		}
	}

	return result, nil
}

func (c *Client) getOrCreateUnderlyingClient() (GogoBoshClient, error) {
	if c.client != nil {
		return c.client, nil
	}

	config := &gogobosh.Config{
		BOSHAddress:       fmt.Sprintf("https://%s:25555", c.Environment),
		ClientID:          c.ClientID,
		ClientSecret:      c.ClientSecret,
		HttpClient:        http.DefaultClient,
		SkipSslValidation: true,
	}

	log.WithoutContext().Debugf("Creating bosh client to connect to %s", config.BOSHAddress)
	client, err := gogobosh.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create bosh client: %w", err)
	}

	c.client = client
	return client, nil
}
