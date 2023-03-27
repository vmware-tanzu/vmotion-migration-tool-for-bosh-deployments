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
	_ = os.Setenv("VCENTER1_PASSWORD", "vcenter1Secret")
	defer func() { _ = os.Unsetenv("VCENTER1_PASSWORD") }()
	_ = os.Setenv("VCENTER2_PASSWORD", "vcenter2Secret")
	defer func() { _ = os.Unsetenv("VCENTER2_PASSWORD") }()
	_ = os.Setenv("BOSH_CLIENT_SECRET", "boshSecret")
	defer func() { _ = os.Unsetenv("BOSH_CLIENT_SECRET") }()

	c, err := config.NewConfigFromFile("./fixtures/config.yml")
	require.NoError(t, err)
	vc1 := &config.VCenter{
		Host:       "sc3-m01-vc01.plat-svcs.pez.vmware.com",
		Username:   "administrator@vsphere.local",
		Password:   "vcenter1Secret",
		Insecure:   true,
		Datacenter: "Datacenter1",
	}
	vc2 := &config.VCenter{
		Host:       "sc3-m01-vc02.plat-svcs.pez.vmware.com",
		Username:   "administrator2@vsphere.local",
		Password:   "vcenter2Secret",
		Insecure:   true,
		Datacenter: "Datacenter2",
	}
	expected := config.Config{
		WorkerPoolSize: 2,
		Compute: config.Compute{
			Source: []config.ComputeAZ{
				{
					Name:    "az1",
					VCenter: vc1,
					Clusters: []config.ComputeCluster{
						{
							Name:         "cf1",
							ResourcePool: "pas-az1",
						},
					},
				},
				{
					Name:    "az2",
					VCenter: vc1,
					Clusters: []config.ComputeCluster{
						{
							Name:         "cf2",
							ResourcePool: "pas-az2",
						},
					},
				},
				{
					Name:    "az3",
					VCenter: vc1,
					Clusters: []config.ComputeCluster{
						{
							Name:         "cf3",
							ResourcePool: "pas-az3",
						},
					},
				},
			},
			Target: []config.ComputeAZ{
				{
					Name:    "az1",
					VCenter: vc2,
					Clusters: []config.ComputeCluster{
						{
							Name:         "tanzu-1",
							ResourcePool: "tas-az1",
						},
					},
				},
				{
					Name:    "az2",
					VCenter: vc2,
					Clusters: []config.ComputeCluster{
						{
							Name:         "tanzu-2",
							ResourcePool: "tas-az2",
						},
					},
				},
				{
					Name:    "az3",
					VCenter: vc2,
					Clusters: []config.ComputeCluster{
						{
							Name:         "tanzu-3",
							ResourcePool: "tas-az3",
						},
					},
				},
			},
		},
		Bosh: &config.Bosh{
			Host:         "10.1.3.12",
			ClientID:     "ops_manager",
			ClientSecret: "boshSecret",
		},
		NetworkMap: map[string]string{
			"PAS-Deployment": "TAS",
			"PAS-Services":   "Services",
		},
		DatastoreMap: map[string]string{
			"ds1": "ssd-ds1",
			"ds2": "ssd-ds2",
		},
		AdditionalVMs: map[string][]string{
			"az1": {
				"vm-2b8bc4a2-90c8-4715-9bc7-ddf64560fdd5",
				"ops-manager-2.10.27",
			},
		},
	}
	require.Equal(t, expected, c)
}

func TestReverseConfig(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/config.yml")
	require.NoError(t, err)
	vc1 := &config.VCenter{
		Host:       "sc3-m01-vc01.plat-svcs.pez.vmware.com",
		Username:   "administrator@vsphere.local",
		Insecure:   true,
		Datacenter: "Datacenter1",
	}
	vc2 := &config.VCenter{
		Host:       "sc3-m01-vc02.plat-svcs.pez.vmware.com",
		Username:   "administrator2@vsphere.local",
		Insecure:   true,
		Datacenter: "Datacenter2",
	}
	expected := config.Config{
		WorkerPoolSize: 2,
		Compute: config.Compute{
			Source: []config.ComputeAZ{
				{
					Name:    "az1",
					VCenter: vc2,
					Clusters: []config.ComputeCluster{
						{
							Name:         "tanzu-1",
							ResourcePool: "tas-az1",
						},
					},
				},
				{
					Name:    "az2",
					VCenter: vc2,
					Clusters: []config.ComputeCluster{
						{
							Name:         "tanzu-2",
							ResourcePool: "tas-az2",
						},
					},
				},
				{
					Name:    "az3",
					VCenter: vc2,
					Clusters: []config.ComputeCluster{
						{
							Name:         "tanzu-3",
							ResourcePool: "tas-az3",
						},
					},
				},
			},
			Target: []config.ComputeAZ{
				{
					Name:    "az1",
					VCenter: vc1,
					Clusters: []config.ComputeCluster{
						{
							Name:         "cf1",
							ResourcePool: "pas-az1",
						},
					},
				},
				{
					Name:    "az2",
					VCenter: vc1,
					Clusters: []config.ComputeCluster{
						{
							Name:         "cf2",
							ResourcePool: "pas-az2",
						},
					},
				},
				{
					Name:    "az3",
					VCenter: vc1,
					Clusters: []config.ComputeCluster{
						{
							Name:         "cf3",
							ResourcePool: "pas-az3",
						},
					},
				},
			},
		},
		Bosh: &config.Bosh{
			Host:     "10.1.3.12",
			ClientID: "ops_manager",
		},
		NetworkMap: map[string]string{
			"TAS":      "PAS-Deployment",
			"Services": "PAS-Services",
		},
		DatastoreMap: map[string]string{
			"ssd-ds1": "ds1",
			"ssd-ds2": "ds2",
		},
		AdditionalVMs: map[string][]string{
			"az1": {
				"vm-2b8bc4a2-90c8-4715-9bc7-ddf64560fdd5",
				"ops-manager-2.10.27",
			},
		},
	}
	rc := c.Reversed()
	require.Equal(t, expected, rc)
}

func TestConfigWithSameTargetVCenter(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/config-same-vcenter.yml")
	require.NoError(t, err)
	vc1 := &config.VCenter{
		Host:       "sc3-m01-vc01.plat-svcs.pez.vmware.com",
		Username:   "administrator@vsphere.local",
		Insecure:   true,
		Datacenter: "Datacenter1",
	}
	expected := config.Config{
		WorkerPoolSize: 2,
		Compute: config.Compute{
			Source: []config.ComputeAZ{
				{
					Name:    "az1",
					VCenter: vc1,
					Clusters: []config.ComputeCluster{
						{
							Name:         "cf1",
							ResourcePool: "pas-az1",
						},
					},
				},
				{
					Name:    "az2",
					VCenter: vc1,
					Clusters: []config.ComputeCluster{
						{
							Name:         "cf2",
							ResourcePool: "pas-az2",
						},
					},
				},
				{
					Name:    "az3",
					VCenter: vc1,
					Clusters: []config.ComputeCluster{
						{
							Name:         "cf3",
							ResourcePool: "pas-az3",
						},
					},
				},
			},
			Target: []config.ComputeAZ{
				{
					Name:    "az1",
					VCenter: vc1,
					Clusters: []config.ComputeCluster{
						{
							Name:         "tanzu-1",
							ResourcePool: "tas-az1",
						},
					},
				},
				{
					Name:    "az2",
					VCenter: vc1,
					Clusters: []config.ComputeCluster{
						{
							Name:         "tanzu-2",
							ResourcePool: "tas-az2",
						},
					},
				},
				{
					Name:    "az3",
					VCenter: vc1,
					Clusters: []config.ComputeCluster{
						{
							Name:         "tanzu-3",
							ResourcePool: "tas-az3",
						},
					},
				},
			},
		},
		Bosh: &config.Bosh{
			Host:     "10.1.3.12",
			ClientID: "ops_manager",
		},
		NetworkMap: map[string]string{
			"PAS-Deployment": "TAS",
			"PAS-Services":   "Services",
		},
		DatastoreMap: map[string]string{
			"ds1": "ssd-ds1",
			"ds2": "ssd-ds2",
		},
	}
	require.Equal(t, expected, c)
}

func TestConfigMinimal(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/config-minimal.yml")
	require.NoError(t, err)
	require.Equal(t, 3, c.WorkerPoolSize)
	require.Equal(t, false, c.DryRun)
	require.Len(t, c.Compute.Source, 1)
	require.Len(t, c.Compute.Target, 1)
	require.Equal(t, c.Compute.Source[0].Name, "az1")
	require.Equal(t, c.Compute.Target[0].Name, "az1")
	require.Len(t, c.Compute.Source[0].Clusters, 1)
	require.Equal(t, c.Compute.Source[0].Clusters[0].Name, "cf1")
	require.Equal(t, c.Compute.Target[0].Clusters[0].Name, "tanzu-1")
	require.Equal(t, c.Compute.Source[0].Clusters[0].ResourcePool, "")
	require.Equal(t, c.Compute.Target[0].Clusters[0].ResourcePool, "")
	require.Equal(t, c.Compute.Source[0].VCenter.Host, "sc3-m01-vc01.plat-svcs.pez.vmware.com")
	require.Equal(t, c.Compute.Target[0].VCenter.Host, "sc3-m01-vc01.plat-svcs.pez.vmware.com")
	require.Len(t, c.DatastoreMap, 1)
	require.Equal(t, c.DatastoreMap["ds1"], "ssd-ds1")
	require.Equal(t, c.NetworkMap["PAS-Deployment"], "TAS")
}

func TestComputeByAZ(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/config.yml")
	require.NoError(t, err)
	tc := c.Compute.TargetByAZ("az1")
	require.Equal(t, "az1", tc.Name)
	require.Len(t, tc.Clusters, 1)
	require.Equal(t, "tanzu-1", tc.Clusters[0].Name)
	require.Equal(t, "tas-az1", tc.Clusters[0].ResourcePool)
	tc = c.Compute.TargetByAZ("az2")
	require.Len(t, tc.Clusters, 1)
	require.Equal(t, "az2", tc.Name)
	require.Equal(t, "tanzu-2", tc.Clusters[0].Name)
	require.Equal(t, "tas-az2", tc.Clusters[0].ResourcePool)
	tc = c.Compute.TargetByAZ("az3")
	require.Len(t, tc.Clusters, 1)
	require.Equal(t, "az3", tc.Name)
	require.Equal(t, "tanzu-3", tc.Clusters[0].Name)
	require.Equal(t, "tas-az3", tc.Clusters[0].ResourcePool)
	tc = c.Compute.TargetByAZ("az-nope")
	require.Nil(t, tc)
}

func TestConfigNoBosh(t *testing.T) {
	c, err := config.NewConfigFromFile("./fixtures/config-no-bosh.yml")
	require.NoError(t, err)
	require.Equal(t, 3, c.WorkerPoolSize)
	require.Equal(t, false, c.DryRun)
	require.Equal(t, "opsmanager", c.AdditionalVMs["az1"][0])
	require.Nil(t, c.Bosh)
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
