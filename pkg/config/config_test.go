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
		Source: config.Source{
			Bosh: config.Bosh{
				Host:     "10.1.3.12",
				ClientID: "ops_manager",
			},
			VCenter: config.VCenter{
				Host:     "sc3-m01-vc01.plat-svcs.pez.vmware.com",
				Username: "administrator@vsphere.local",
				Insecure: false,
			},
			Datacenter: "Datacenter1",
		},
		Target: config.Target{
			VCenter: config.VCenter{
				Host:     "sc3-m01-vc02.plat-svcs.pez.vmware.com",
				Username: "administrator2@vsphere.local",
				Insecure: false,
			},
			Datastore:  "nfs01",
			Cluster:    "Cluster01",
			Datacenter: "Datacenter2",
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
	}
	require.Equal(t, expected, c)
}

func TestConfigWithSameTargetVCenter(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/config-same-vcenter.yml")
	require.NoError(t, err)
	expected := config.Config{
		Source: config.Source{
			Bosh: config.Bosh{
				Host:     "10.1.3.12",
				ClientID: "ops_manager",
			},
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
			Datastore:  "nfs01",
			Cluster:    "Cluster01",
			Datacenter: "Datacenter2",
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
	}
	require.Equal(t, expected, c)
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
