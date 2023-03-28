/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package migrate_test

import (
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/bosh"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/config"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate"
	"testing"
)

func baseConfig() config.Config {
	return config.Config{
		Bosh: &config.Bosh{
			Host:         "192.168.1.2",
			ClientID:     "admin",
			ClientSecret: "secret",
		},
		DryRun:         false,
		WorkerPoolSize: 1,
		NetworkMap: map[string]string{
			"Net1": "Net2",
		},
		DatastoreMap: map[string]string{
			"DS1": "DS2",
		},
		Compute: config.Compute{
			Source: []config.ComputeAZ{
				{
					Name: "az1",
					VCenter: &config.VCenter{
						Host:       "vcenter1.example.com",
						Username:   "admin1",
						Password:   "secret1",
						Datacenter: "DC1",
					},
					Clusters: []config.ComputeCluster{
						{
							Name:         "Cluster1",
							ResourcePool: "RP1",
						},
					},
				},
			},
			Target: []config.ComputeAZ{
				{
					Name: "az1",
					VCenter: &config.VCenter{
						Host:       "vcenter2.example.com",
						Username:   "admin2",
						Password:   "secret2",
						Datacenter: "DC2",
					},
					Clusters: []config.ComputeCluster{
						{
							Name:         "Cluster2",
							ResourcePool: "RP2",
						},
					},
				},
			},
		},
		AdditionalVMs: map[string][]string{
			"az1": {
				"additional-vm1",
			},
		},
	}
}

func TestNewFoundationMigratorFromConfig(t *testing.T) {
	_, err := migrate.NewFoundationMigratorFromConfig(baseConfig())
	require.NoError(t, err)
}

func TestConfigToBoshClient(t *testing.T) {
	c := baseConfig()
	b := migrate.ConfigToBoshClient(c).(*bosh.Client)
	require.Equal(t, "192.168.1.2", b.Environment)
	require.Equal(t, "admin", b.ClientID)
	require.Equal(t, "secret", b.ClientSecret)

	c.Bosh = nil
	bf := migrate.ConfigToBoshClient(c)
	require.IsType(t, migrate.NullBoshClient{}, bf)
}

func TestConfigToVCenterClientPool(t *testing.T) {
	c := baseConfig()
	c.Compute.Source = append(c.Compute.Source, config.ComputeAZ{
		Name: "az2",
		VCenter: &config.VCenter{
			Host:       "vcenter3.example.com",
			Username:   "admin3",
			Password:   "secret3",
			Datacenter: "DC3",
		},
		Clusters: []config.ComputeCluster{
			{
				Name:         "Cluster3",
				ResourcePool: "RP3",
			},
		},
	})
	c.Compute.Target = append(c.Compute.Target, config.ComputeAZ{
		Name: "az2",
		VCenter: &config.VCenter{
			Host:       "vcenter4.example.com",
			Username:   "admin4",
			Password:   "secret4",
			Datacenter: "DC4",
		},
		Clusters: []config.ComputeCluster{
			{
				Name:         "Cluster4",
				ResourcePool: "RP4",
			},
		},
	})

	p := migrate.ConfigToVCenterClientPool(c)
	require.Contains(t, p.SourceAZs(), "az1")
	require.Contains(t, p.SourceAZs(), "az2")
	require.Contains(t, p.TargetAZs(), "az1")
	require.Contains(t, p.TargetAZs(), "az2")

	sc1 := p.GetSourceClientByAZ("az1")
	require.NotNil(t, sc1)
	require.Equal(t, "vcenter1.example.com", sc1.Host)
	require.Equal(t, "admin1", sc1.Username)
	require.Equal(t, "secret1", sc1.Password)

	sc2 := p.GetSourceClientByAZ("az2")
	require.NotNil(t, sc2)
	require.Equal(t, "vcenter3.example.com", sc2.Host)
	require.Equal(t, "admin3", sc2.Username)
	require.Equal(t, "secret3", sc2.Password)

	tc1 := p.GetTargetClientByAZ("az1")
	require.NotNil(t, tc1)
	require.Equal(t, "vcenter2.example.com", tc1.Host)
	require.Equal(t, "admin2", tc1.Username)
	require.Equal(t, "secret2", tc1.Password)

	tc2 := p.GetTargetClientByAZ("az2")
	require.NotNil(t, tc2)
	require.Equal(t, "vcenter4.example.com", tc2.Host)
	require.Equal(t, "admin4", tc2.Username)
	require.Equal(t, "secret4", tc2.Password)
}

func TestOneToOneClusterAZMapping(t *testing.T) {
	c := baseConfig()
	azMapping, err := migrate.ConfigToAZMapping(c)
	require.NoError(t, err)
	require.Len(t, azMapping, 1)

	s := azMapping[0].Source
	d := azMapping[0].Target

	require.Equal(t, "az1", s.Name)
	require.Equal(t, "az1", d.Name)
	require.Equal(t, "DC1", s.Datacenter)
	require.Equal(t, "DC2", d.Datacenter)
	require.Equal(t, "Cluster1", s.Cluster)
	require.Equal(t, "Cluster2", d.Cluster)
	require.Equal(t, "RP1", s.ResourcePool)
	require.Equal(t, "RP2", d.ResourcePool)

	hpc := migrate.ConfigToTargetHostPoolConfig(c)
	require.NotNil(t, hpc)
	require.Len(t, hpc.AZs, 1)
	require.Contains(t, hpc.AZs, "az1")
	require.Len(t, hpc.AZs["az1"].Clusters, 1)
	require.Equal(t, "Cluster2", hpc.AZs["az1"].Clusters[0])
}

func TestOneToManyClustersAZMapping(t *testing.T) {
	c := baseConfig()
	c.Compute.Target[0].Clusters = append(c.Compute.Target[0].Clusters, config.ComputeCluster{
		Name:         "Cluster3",
		ResourcePool: "RP3",
	})
	azMapping, err := migrate.ConfigToAZMapping(c)
	require.NoError(t, err)
	require.Len(t, azMapping, 2)

	s1 := azMapping[0].Source
	s2 := azMapping[1].Source
	d1 := azMapping[0].Target
	d2 := azMapping[1].Target

	require.Equal(t, "az1", s1.Name)
	require.Equal(t, "az1", d1.Name)
	require.Equal(t, "DC1", s1.Datacenter)
	require.Equal(t, "DC2", d1.Datacenter)
	require.Equal(t, "Cluster1", s1.Cluster)
	require.Equal(t, "Cluster2", d1.Cluster)
	require.Equal(t, "RP1", s1.ResourcePool)
	require.Equal(t, "RP2", d1.ResourcePool)

	require.Equal(t, "az1", s2.Name)
	require.Equal(t, "az1", d2.Name)
	require.Equal(t, "DC1", s2.Datacenter)
	require.Equal(t, "DC2", d2.Datacenter)
	require.Equal(t, "Cluster1", s2.Cluster)
	require.Equal(t, "Cluster3", d2.Cluster)
	require.Equal(t, "RP1", s2.ResourcePool)
	require.Equal(t, "RP3", d2.ResourcePool)

	hpc := migrate.ConfigToTargetHostPoolConfig(c)
	require.NotNil(t, hpc)
	require.Len(t, hpc.AZs, 1)
	require.Contains(t, hpc.AZs, "az1")
	require.Len(t, hpc.AZs["az1"].Clusters, 2)
	require.Equal(t, "Cluster2", hpc.AZs["az1"].Clusters[0])
	require.Equal(t, "Cluster3", hpc.AZs["az1"].Clusters[1])
}

func TestManyToOneClustersAZMapping(t *testing.T) {
	c := baseConfig()
	c.Compute.Source[0].Clusters = append(c.Compute.Source[0].Clusters, config.ComputeCluster{
		Name:         "Cluster3",
		ResourcePool: "RP3",
	})
	azMapping, err := migrate.ConfigToAZMapping(c)
	require.NoError(t, err)
	require.Len(t, azMapping, 2)

	s1 := azMapping[0].Source
	s2 := azMapping[1].Source
	d1 := azMapping[0].Target
	d2 := azMapping[1].Target

	require.Equal(t, "az1", s1.Name)
	require.Equal(t, "az1", d1.Name)
	require.Equal(t, "DC1", s1.Datacenter)
	require.Equal(t, "DC2", d1.Datacenter)
	require.Equal(t, "Cluster1", s1.Cluster)
	require.Equal(t, "Cluster2", d1.Cluster)
	require.Equal(t, "RP1", s1.ResourcePool)
	require.Equal(t, "RP2", d1.ResourcePool)

	require.Equal(t, "az1", s2.Name)
	require.Equal(t, "az1", d2.Name)
	require.Equal(t, "DC1", s2.Datacenter)
	require.Equal(t, "DC2", d2.Datacenter)
	require.Equal(t, "Cluster3", s2.Cluster)
	require.Equal(t, "Cluster2", d2.Cluster)
	require.Equal(t, "RP3", s2.ResourcePool)
	require.Equal(t, "RP2", d2.ResourcePool)

	hpc := migrate.ConfigToTargetHostPoolConfig(c)
	require.NotNil(t, hpc)
	require.Len(t, hpc.AZs, 1)
	require.Contains(t, hpc.AZs, "az1")
	require.Len(t, hpc.AZs["az1"].Clusters, 1)
	require.Equal(t, "Cluster2", hpc.AZs["az1"].Clusters[0])
}

func TestManyToManyClustersAZMapping(t *testing.T) {
	c := baseConfig()
	c.Compute.Source[0].Clusters = append(c.Compute.Source[0].Clusters, config.ComputeCluster{
		Name:         "Cluster3",
		ResourcePool: "RP3",
	})
	c.Compute.Target[0].Clusters = append(c.Compute.Target[0].Clusters, config.ComputeCluster{
		Name:         "Cluster4",
		ResourcePool: "RP4",
	})
	azMapping, err := migrate.ConfigToAZMapping(c)
	require.NoError(t, err)
	require.Len(t, azMapping, 4)

	s1 := azMapping[0].Source
	s2 := azMapping[1].Source
	s3 := azMapping[2].Source
	s4 := azMapping[3].Source
	d1 := azMapping[0].Target
	d2 := azMapping[1].Target
	d3 := azMapping[2].Target
	d4 := azMapping[3].Target

	require.Equal(t, "az1", s1.Name)
	require.Equal(t, "az1", d1.Name)
	require.Equal(t, "DC1", s1.Datacenter)
	require.Equal(t, "DC2", d1.Datacenter)
	require.Equal(t, "Cluster1", s1.Cluster)
	require.Equal(t, "Cluster2", d1.Cluster)
	require.Equal(t, "RP1", s1.ResourcePool)
	require.Equal(t, "RP2", d1.ResourcePool)

	require.Equal(t, "az1", s2.Name)
	require.Equal(t, "az1", d2.Name)
	require.Equal(t, "DC1", s2.Datacenter)
	require.Equal(t, "DC2", d2.Datacenter)
	require.Equal(t, "Cluster1", s2.Cluster)
	require.Equal(t, "Cluster4", d2.Cluster)
	require.Equal(t, "RP1", s2.ResourcePool)
	require.Equal(t, "RP4", d2.ResourcePool)

	require.Equal(t, "az1", s3.Name)
	require.Equal(t, "az1", d3.Name)
	require.Equal(t, "DC1", s3.Datacenter)
	require.Equal(t, "DC2", d3.Datacenter)
	require.Equal(t, "Cluster3", s3.Cluster)
	require.Equal(t, "Cluster2", d3.Cluster)
	require.Equal(t, "RP3", s3.ResourcePool)
	require.Equal(t, "RP2", d3.ResourcePool)

	require.Equal(t, "az1", s4.Name)
	require.Equal(t, "az1", d4.Name)
	require.Equal(t, "DC1", s4.Datacenter)
	require.Equal(t, "DC2", d4.Datacenter)
	require.Equal(t, "Cluster3", s4.Cluster)
	require.Equal(t, "Cluster4", d4.Cluster)
	require.Equal(t, "RP3", s4.ResourcePool)
	require.Equal(t, "RP4", d4.ResourcePool)

	hpc := migrate.ConfigToTargetHostPoolConfig(c)
	require.NotNil(t, hpc)
	require.Len(t, hpc.AZs, 1)
	require.Contains(t, hpc.AZs, "az1")
	require.Len(t, hpc.AZs["az1"].Clusters, 2)
	require.Equal(t, "Cluster2", hpc.AZs["az1"].Clusters[0])
	require.Equal(t, "Cluster4", hpc.AZs["az1"].Clusters[1])
}

func TestOneToOneClusterMultipleAZsMapping(t *testing.T) {
	c := baseConfig()
	c.Compute.Source = append(c.Compute.Source, config.ComputeAZ{
		Name: "az2",
		VCenter: &config.VCenter{
			Host:       "vcenter1.example.com",
			Username:   "admin1",
			Password:   "secret1",
			Datacenter: "DC1",
		},
		Clusters: []config.ComputeCluster{
			{
				Name:         "Cluster3",
				ResourcePool: "RP3",
			},
		},
	})
	c.Compute.Target = append(c.Compute.Target, config.ComputeAZ{
		Name: "az2",
		VCenter: &config.VCenter{
			Host:       "vcenter2.example.com",
			Username:   "admin2",
			Password:   "secret2",
			Datacenter: "DC2",
		},
		Clusters: []config.ComputeCluster{
			{
				Name:         "Cluster4",
				ResourcePool: "RP4",
			},
		},
	})

	azMapping, err := migrate.ConfigToAZMapping(c)
	require.NoError(t, err)
	require.Len(t, azMapping, 2)

	s1 := azMapping[0].Source
	d1 := azMapping[0].Target
	s2 := azMapping[1].Source
	d2 := azMapping[1].Target

	require.Equal(t, "az1", s1.Name)
	require.Equal(t, "az1", d1.Name)
	require.Equal(t, "DC1", s1.Datacenter)
	require.Equal(t, "DC2", d1.Datacenter)
	require.Equal(t, "Cluster1", s1.Cluster)
	require.Equal(t, "Cluster2", d1.Cluster)
	require.Equal(t, "RP1", s1.ResourcePool)
	require.Equal(t, "RP2", d1.ResourcePool)

	require.Equal(t, "az2", s2.Name)
	require.Equal(t, "az2", d2.Name)
	require.Equal(t, "DC1", s2.Datacenter)
	require.Equal(t, "DC2", d2.Datacenter)
	require.Equal(t, "Cluster3", s2.Cluster)
	require.Equal(t, "Cluster4", d2.Cluster)
	require.Equal(t, "RP3", s2.ResourcePool)
	require.Equal(t, "RP4", d2.ResourcePool)

	hpc := migrate.ConfigToTargetHostPoolConfig(c)
	require.NotNil(t, hpc)
	require.Len(t, hpc.AZs, 2)
	require.Contains(t, hpc.AZs, "az1")
	require.Len(t, hpc.AZs["az1"].Clusters, 1)
	require.Equal(t, "Cluster2", hpc.AZs["az1"].Clusters[0])

	require.Contains(t, hpc.AZs, "az2")
	require.Len(t, hpc.AZs["az2"].Clusters, 1)
	require.Equal(t, "Cluster4", hpc.AZs["az2"].Clusters[0])
}
