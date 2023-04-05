/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package bosh_test

import (
	"context"
	"errors"
	"github.com/cloudfoundry-community/gogobosh"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/bosh"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/bosh/boshfakes"
	"testing"
)

func TestVMsAndStemcells_ReturnsErrorWhenCloudConfigNotFound(t *testing.T) {
	configs := []gogobosh.Cfg{
		{
			ID:      "7",
			Name:    "runtime",
			Type:    "director_runtime",
			Content: "",
		},
		{
			ID:      "1",
			Name:    "cloud",
			Type:    "pivotal-container-service-0fed31819f77fbf3f90b",
			Content: "",
		},
	}

	gb := &boshfakes.FakeGogoBoshClient{}
	gb.GetCloudConfigReturns(configs, nil)

	c := bosh.NewFromGogoBoshClient(gb)
	_, err := c.VMsAndStemcells(context.Background())
	require.Error(t, err)
	require.Equal(t, "could not find BOSH default cloud config", err.Error())
}

func TestVMsAndStemcells_ReturnsErrorWhenCloudConfigNotYaml(t *testing.T) {
	configs := []gogobosh.Cfg{
		{
			ID:      "6",
			Name:    "default",
			Type:    "cloud",
			Content: "garbage",
		},
	}

	gb := &boshfakes.FakeGogoBoshClient{}
	gb.GetCloudConfigReturns(configs, nil)

	c := bosh.NewFromGogoBoshClient(gb)
	_, err := c.VMsAndStemcells(context.Background())
	require.ErrorContains(t, err, "could not unmarshal BOSH cloud config: yaml")
}

func TestVMsAndStemcells_ReturnsErrorWhenStemcellCPINotFound(t *testing.T) {
	configs := []gogobosh.Cfg{
		{
			ID:      "6",
			Name:    "default",
			Type:    "cloud",
			Content: cloudConfigYaml,
		},
	}

	gb := &boshfakes.FakeGogoBoshClient{}
	gb.GetCloudConfigReturns(configs, nil)

	sc := []gogobosh.Stemcell{
		{
			Name:            "bosh-vsphere-esxi-ubuntu-jammy-go_agent",
			OperatingSystem: "ubuntu-jammy",
			Version:         "1.93",
			CID:             "sc-guid",
			CPI:             "does-not-exist",
		},
	}
	gb.GetStemcellsReturns(sc, nil)

	c := bosh.NewFromGogoBoshClient(gb)
	_, err := c.VMsAndStemcells(context.Background())
	require.ErrorContains(t, err, "could not find a CPI to AZ mapping for stemcell sc-guid")
}

func TestVMsAndStemcells_ReturnsErrorWhenDeploymentVMsNotFound(t *testing.T) {
	configs := []gogobosh.Cfg{
		{
			ID:      "6",
			Name:    "default",
			Type:    "cloud",
			Content: cloudConfigYaml,
		},
	}

	gb := &boshfakes.FakeGogoBoshClient{}
	gb.GetCloudConfigReturns(configs, nil)

	sc := []gogobosh.Stemcell{
		{
			Name:            "bosh-vsphere-esxi-ubuntu-jammy-go_agent",
			OperatingSystem: "ubuntu-jammy",
			Version:         "1.93",
			CID:             "sc-guid",
			CPI:             "1e668fac900079c31a44",
		},
	}
	gb.GetStemcellsReturns(sc, nil)

	d := []gogobosh.Deployment{
		{
			Name:        "pivotal-container-service-guid",
			CloudConfig: "latest",
		},
	}
	gb.GetDeploymentsReturns(d, nil)
	gb.GetDeploymentVMsReturns(nil, errors.New("could not get VMs for deployment"))

	c := bosh.NewFromGogoBoshClient(gb)
	_, err := c.VMsAndStemcells(context.Background())
	require.ErrorContains(t, err, "failed to get deployment pivotal-container-service-guid VMs:")
}

func TestVMsAndStemcells_ReturnsStemcellsAndVMs(t *testing.T) {
	configs := []gogobosh.Cfg{
		{
			ID:      "6",
			Name:    "default",
			Type:    "cloud",
			Content: cloudConfigYaml,
		},
	}

	gb := &boshfakes.FakeGogoBoshClient{}
	gb.GetCloudConfigReturns(configs, nil)

	sc := []gogobosh.Stemcell{
		{
			Name:            "bosh-vsphere-esxi-ubuntu-jammy-go_agent",
			OperatingSystem: "ubuntu-jammy",
			Version:         "1.93",
			CID:             "sc1-guid",
			CPI:             "1e668fac900079c31a44",
		},
		{
			Name:            "bosh-vsphere-esxi-ubuntu-jammy-go_agent",
			OperatingSystem: "ubuntu-jammy",
			Version:         "1.93",
			CID:             "sc2-guid",
			CPI:             "622c5ec3e67dde98d08c",
		},
	}
	gb.GetStemcellsReturns(sc, nil)

	d := []gogobosh.Deployment{
		{
			Name:        "pivotal-container-service-guid",
			CloudConfig: "latest",
		},
	}
	gb.GetDeploymentsReturns(d, nil)

	v := []gogobosh.VM{
		{
			VMCID:   "vm-guid1",
			IPs:     []string{"192.168.2.11"},
			AgentID: "agent1-guid",
			JobName: "pks-db",
			Index:   0,
			AZ:      "az1",
			ID:      "id1-guid",
		},
		{
			VMCID:   "vm-guid2",
			IPs:     []string{"192.168.2.12"},
			AgentID: "agent2-guid",
			JobName: "pks-api",
			Index:   0,
			AZ:      "az2",
			ID:      "id2-guid",
		},
	}
	gb.GetDeploymentVMsReturns(v, nil)

	c := bosh.NewFromGogoBoshClient(gb)
	vms, err := c.VMsAndStemcells(context.Background())
	require.NoError(t, err)
	require.Len(t, vms, 4)

	require.Equal(t, "sc1-guid", vms[0].Name)
	require.Equal(t, "az1", vms[0].AZ)

	require.Equal(t, "sc2-guid", vms[1].Name)
	require.Equal(t, "az2", vms[1].AZ)

	require.Equal(t, "vm-guid1", vms[2].Name)
	require.Equal(t, "az1", vms[2].AZ)

	require.Equal(t, "vm-guid2", vms[3].Name)
	require.Equal(t, "az2", vms[3].AZ)
}

const cloudConfigYaml = `
azs:
- cpi: 1e668fac900079c31a44
  name: az1
  cloud_properties:
    datacenters:
    - clusters:
      - vc01cl01:
          resource_pool: tkgi1
      name: az1
- name: az2
  cpi: 622c5ec3e67dde98d08c
  cloud_properties:
    datacenters:
    - name: az2
      clusters:
      - vc01cl01:
          resource_pool: 
`
