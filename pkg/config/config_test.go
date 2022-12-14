/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/config"
)

func TestConfig(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/config.yml")
	require.NoError(t, err)
	expected := config.Config{
		WorkerPoolSize: 2,
		Source: config.Source{
			VCenter: config.VCenter{
				Host:     "sc3-m01-vc01.plat-svcs.pez.vmware.com",
				Username: "administrator@vsphere.local",
				Insecure: true,
			},
			Datacenter: "Datacenter1",
		},
		Target: config.Target{
			VCenter: config.VCenter{
				Host:     "sc3-m01-vc02.plat-svcs.pez.vmware.com",
				Username: "administrator2@vsphere.local",
				Insecure: true,
			},
			Datacenter: "Datacenter2",
		},
		Bosh: &config.Bosh{
			Host:     "10.1.3.12",
			ClientID: "ops_manager",
		},
		NetworkMap: map[string]string{
			"PAS-Deployment": "TAS",
			"PAS-Services":   "Services",
		},
		ResourcePoolMap: map[string]string{
			"pas-az1": "tas-az1",
			"pas-az2": "tas-az2",
			"pas-az3": "tas-az3",
		},
		DatastoreMap: map[string]string{
			"ds1": "ssd-ds1",
			"ds2": "ssd-ds2",
		},
		ClusterMap: map[string]string{
			"cf1": "tanzu-1",
			"cf2": "tanzu-2",
			"cf3": "tanzu-3",
		},
		AdditionalVMs: []string{
			"vm-2b8bc4a2-90c8-4715-9bc7-ddf64560fdd5",
			"ops-manager-2.10.27",
		},
	}
	require.Equal(t, expected, c)
}

func TestReverseConfig(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/config.yml")
	require.NoError(t, err)
	expected := config.Config{
		WorkerPoolSize: 2,
		Target: config.Target{
			VCenter: config.VCenter{
				Host:     "sc3-m01-vc01.plat-svcs.pez.vmware.com",
				Username: "administrator@vsphere.local",
				Insecure: true,
			},
			Datacenter: "Datacenter1",
		},
		Source: config.Source{
			VCenter: config.VCenter{
				Host:     "sc3-m01-vc02.plat-svcs.pez.vmware.com",
				Username: "administrator2@vsphere.local",
				Insecure: true,
			},
			Datacenter: "Datacenter2",
		},
		Bosh: &config.Bosh{
			Host:     "10.1.3.12",
			ClientID: "ops_manager",
		},
		NetworkMap: map[string]string{
			"TAS":      "PAS-Deployment",
			"Services": "PAS-Services",
		},
		ResourcePoolMap: map[string]string{
			"tas-az1": "pas-az1",
			"tas-az2": "pas-az2",
			"tas-az3": "pas-az3",
		},
		DatastoreMap: map[string]string{
			"ssd-ds1": "ds1",
			"ssd-ds2": "ds2",
		},
		ClusterMap: map[string]string{
			"tanzu-1": "cf1",
			"tanzu-2": "cf2",
			"tanzu-3": "cf3",
		},
		AdditionalVMs: []string{
			"vm-2b8bc4a2-90c8-4715-9bc7-ddf64560fdd5",
			"ops-manager-2.10.27",
		},
	}
	rc := c.Reversed()
	require.Equal(t, expected, rc)
}

func TestConfigWithSameTargetVCenter(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/config-same-vcenter.yml")
	require.NoError(t, err)
	expected := config.Config{
		WorkerPoolSize: 2,
		Source: config.Source{
			VCenter: config.VCenter{
				Host:     "sc3-m01-vc01.plat-svcs.pez.vmware.com",
				Username: "administrator@vsphere.local",
				Insecure: false,
			},
			Datacenter: "Datacenter1",
		},
		Target: config.Target{
			VCenter: config.VCenter{
				Host:     "sc3-m01-vc01.plat-svcs.pez.vmware.com",
				Username: "administrator@vsphere.local",
				Insecure: false,
			},
			Datacenter: "Datacenter2",
		},
		Bosh: &config.Bosh{
			Host:     "10.1.3.12",
			ClientID: "ops_manager",
		},
		NetworkMap: map[string]string{
			"PAS-Deployment": "TAS",
			"PAS-Services":   "Services",
		},
		ResourcePoolMap: map[string]string{
			"pas-az1": "tas-az1",
			"pas-az2": "tas-az2",
			"pas-az3": "tas-az3",
		},
		DatastoreMap: map[string]string{
			"ds1": "ssd-ds1",
			"ds2": "ssd-ds2",
		},
		ClusterMap: map[string]string{
			"cf1": "tanzu-1",
			"cf2": "tanzu-2",
			"cf3": "tanzu-3",
		},
	}
	require.Equal(t, expected, c)
}

func TestConfigMinimal(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/config-minimal.yml")
	require.NoError(t, err)
	require.Equal(t, 3, c.WorkerPoolSize)
	require.Equal(t, false, c.DryRun)
	require.Len(t, c.ResourcePoolMap, 0)
}

func TestConfigNoBosh(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/config-no-bosh.yml")
	require.NoError(t, err)
	require.Equal(t, 3, c.WorkerPoolSize)
	require.Equal(t, false, c.DryRun)
	require.Equal(t, "opsmanager", c.AdditionalVMs[0])
	require.Nil(t, c.Bosh)
	require.Len(t, c.ResourcePoolMap, 0)
}

func TestConfigFromMarshalledFile(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/config.yml")
	require.NoError(t, err)
	buf, _ := os.ReadFile("./fixtures/config_marshalled.yml")
	require.Equal(t, string(buf[:]), c.String())
}

func TestConfigFileMissing(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/doesnotexist.yml")
	require.Error(t, err)
	require.Equal(t, config.Config{}, c)
}

func TestConfigFileInvalid(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/bogus.yml")
	require.Error(t, err)
	require.Equal(t, config.Config{}, c)
}
