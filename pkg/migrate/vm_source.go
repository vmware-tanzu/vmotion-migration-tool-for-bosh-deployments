/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package migrate

import (
	"context"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/bosh"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/config"
)

//counterfeiter:generate . BoshClient
type BoshClient interface {
	VMsAndStemcells(context.Context) ([]bosh.VM, error)
}

// NullBoshClient is a null object pattern when no bosh client is specified in the config
type NullBoshClient struct{}

// VMsAndStemcells returns an empty list
func (c NullBoshClient) VMsAndStemcells(context.Context) ([]bosh.VM, error) {
	return []bosh.VM{}, nil
}

type VM struct {
	Name string
	AZ   string

	// list of clusters within the source AZ that may contain the VM
	Clusters []string
}

type VMSource struct {
	BoshClient BoshClient

	additionalVMs    []VM
	srcAZsToClusters map[string][]string
}

func NewVMSourceFromConfig(c config.Config) *VMSource {
	azToClusters := configToSourceClustersByAZ(c)
	additionalVMs := configToAdditionalVMs(c, azToClusters)
	boshClient := configToBoshClient(c)
	return &VMSource{
		BoshClient:       boshClient,
		additionalVMs:    additionalVMs,
		srcAZsToClusters: azToClusters,
	}
}

// VMsToMigrate returns the list of all BOSH and additional VMs to migrate
func (s *VMSource) VMsToMigrate(ctx context.Context) ([]VM, error) {
	boshVMs, err := s.BoshClient.VMsAndStemcells(ctx)
	if err != nil {
		return nil, err
	}

	var vms []VM
	for _, bvm := range boshVMs {
		vms = append(vms, VM{
			Name:     bvm.Name,
			AZ:       bvm.AZ,
			Clusters: s.srcAZsToClusters[bvm.AZ],
		})
	}
	vms = append(vms, s.additionalVMs...)
	return vms, nil
}

func configToAdditionalVMs(c config.Config, srcAZToClusters map[string][]string) []VM {
	var additionalVMs []VM
	for az, vms := range c.AdditionalVMs {
		for _, v := range vms {
			additionalVMs = append(additionalVMs, VM{
				Name:     v,
				AZ:       az,
				Clusters: srcAZToClusters[az],
			})
		}
	}
	return additionalVMs
}

func configToSourceClustersByAZ(c config.Config) map[string][]string {
	azToClusters := map[string][]string{}
	for _, az := range c.Compute.Source {
		var cls []string
		for _, cl := range az.Clusters {
			cls = append(cls, cl.Name)
		}
		azToClusters[az.Name] = cls
	}
	return azToClusters
}

func configToBoshClient(c config.Config) BoshClient {
	// if there's a configured optional BOSH config section then create a client
	if c.Bosh != nil {
		return bosh.New(c.Bosh.Host, c.Bosh.ClientID, c.Bosh.ClientSecret)
	}
	return NullBoshClient{}
}
