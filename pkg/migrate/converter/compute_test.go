/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter_test

import (
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate/converter"
	"strings"
	"testing"
)

func TestTargetComputeWithEmptyArgs(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	_, err := cm.TargetComputeFromSourceAZ(converter.AZ{
		Cluster: "sC1",
		Name:    "AZ1",
	})
	require.Error(t, err, "expected datacenter to be non-empty string")

	cm = converter.NewEmptyMappedCompute()
	_, err = cm.TargetComputeFromSourceAZ(converter.AZ{
		Datacenter: "sDC1",
		Name:       "AZ1",
	})
	require.Error(t, err, "expected cluster to be non-empty string")

	cm = converter.NewEmptyMappedCompute()
	_, err = cm.TargetComputeFromSourceAZ(converter.AZ{
		Datacenter: "sDC1",
		Cluster:    "sC1",
	})
	require.Error(t, err, "expected AZ name to be non-empty string")
}

func TestTargetComputeOneToOneClusterSameVCenter(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZ{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	}, converter.AZ{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	})

	result, err := cm.TargetComputeFromSourceAZ(converter.AZ{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZ{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	}, result)
}

func TestTargetComputeOneToOneClusterDifferentVCenter(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZ{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	}, converter.AZ{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	})

	result, err := cm.TargetComputeFromSourceAZ(converter.AZ{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZ{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	}, result)
}

func TestTargetComputeWithDefaultResourcePool(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZ{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	}, converter.AZ{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	})

	result, err := cm.TargetComputeFromSourceAZ(converter.AZ{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		Name:         "AZ1",
		ResourcePool: "Resources",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZ{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	}, result)
}

func TestTargetComputeOneToOneClusterAndRP(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZ{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		Name:         "AZ1",
	}, converter.AZ{
		Datacenter:   "tDC1",
		Cluster:      "tC1",
		ResourcePool: "tRP1",
		Name:         "AZ1",
	})

	result, err := cm.TargetComputeFromSourceAZ(converter.AZ{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZ{
		Datacenter:   "tDC1",
		Cluster:      "tC1",
		ResourcePool: "tRP1",
		Name:         "AZ1",
	}, result)
}

func TestTargetComputeOneToOneClusterAndOneSrcRP(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZ{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		Name:         "AZ1",
	}, converter.AZ{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	})

	result, err := cm.TargetComputeFromSourceAZ(converter.AZ{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZ{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	}, result)
}

func TestTargetComputeOneToOneClusterAndOneTargetRP(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZ{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	}, converter.AZ{
		Datacenter:   "tDC1",
		Cluster:      "tC1",
		ResourcePool: "tRP1",
		Name:         "AZ1",
	})

	result, err := cm.TargetComputeFromSourceAZ(converter.AZ{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZ{
		Datacenter:   "tDC1",
		Cluster:      "tC1",
		ResourcePool: "tRP1",
		Name:         "AZ1",
	}, result)
}

func TestTargetOneClusterAndRPsToThreeClustersOnDifferentVCenters(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZ{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		Name:         "AZ1",
	}, converter.AZ{
		Datacenter:   "tDC1",
		Cluster:      "tC1",
		ResourcePool: "ResourcePool",
		Name:         "AZ1",
	})
	cm.Add(converter.AZ{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP2",
		Name:         "AZ1",
	}, converter.AZ{
		Datacenter:   "tDC1",
		Cluster:      "tC2",
		ResourcePool: "ResourcePool",
		Name:         "AZ2",
	})
	cm.Add(converter.AZ{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP3",
		Name:         "AZ1",
	}, converter.AZ{
		Datacenter:   "tDC1",
		Cluster:      "tC3",
		ResourcePool: "ResourcePool",
		Name:         "AZ3",
	})

	result, err := cm.TargetComputeFromSourceAZ(converter.AZ{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZ{
		Datacenter:   "tDC1",
		Cluster:      "tC1",
		ResourcePool: "ResourcePool",
		Name:         "AZ1",
	}, result)

	result, err = cm.TargetComputeFromSourceAZ(converter.AZ{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP2",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZ{
		Datacenter:   "tDC1",
		Cluster:      "tC2",
		ResourcePool: "ResourcePool",
		Name:         "AZ2",
	}, result)

	result, err = cm.TargetComputeFromSourceAZ(converter.AZ{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP3",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZ{
		Datacenter:   "tDC1",
		Cluster:      "tC3",
		ResourcePool: "ResourcePool",
		Name:         "AZ3",
	}, result)
}

func TestTargetManyToManyClustersInAZ(t *testing.T) {
	// create many-to-many mapping
	// sc1 => tc1
	// sc1 => tc2
	// sc2 => tc1
	// sc2 => tc2
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZ{
		Datacenter:   "sDC",
		Cluster:      "sCluster1",
		ResourcePool: "sResourcePool",
		Name:         "AZ1",
	}, converter.AZ{
		Datacenter:   "tDC",
		Cluster:      "tCluster1",
		ResourcePool: "tResourcePool",
		Name:         "AZ1",
	})
	cm.Add(converter.AZ{
		Datacenter:   "sDC",
		Cluster:      "sCluster1",
		ResourcePool: "sResourcePool",
		Name:         "AZ1",
	}, converter.AZ{
		Datacenter:   "tDC",
		Cluster:      "tCluster2",
		ResourcePool: "tResourcePool",
		Name:         "AZ1",
	})
	cm.Add(converter.AZ{
		Datacenter:   "sDC",
		Cluster:      "sCluster2",
		ResourcePool: "sResourcePool",
		Name:         "AZ1",
	}, converter.AZ{
		Datacenter:   "tDC",
		Cluster:      "tCluster1",
		ResourcePool: "tResourcePool",
		Name:         "AZ1",
	})
	cm.Add(converter.AZ{
		Datacenter:   "sDC",
		Cluster:      "sCluster2",
		ResourcePool: "sResourcePool",
		Name:         "AZ1",
	}, converter.AZ{
		Datacenter:   "tDC",
		Cluster:      "tCluster2",
		ResourcePool: "tResourcePool",
		Name:         "AZ1",
	})

	// get the first mapping for sc1
	result, err := cm.TargetComputesFromSourceAZ(converter.AZ{
		Datacenter:   "sDC",
		Cluster:      "sCluster1",
		ResourcePool: "sResourcePool",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, converter.AZ{
		Datacenter:   "tDC",
		Cluster:      "tCluster1",
		ResourcePool: "tResourcePool",
		Name:         "AZ1",
	}, result[0])
	require.Equal(t, converter.AZ{
		Datacenter:   "tDC",
		Cluster:      "tCluster2",
		ResourcePool: "tResourcePool",
		Name:         "AZ1",
	}, result[1])

	// get the first mapping for sc2
	result, err = cm.TargetComputesFromSourceAZ(converter.AZ{
		Datacenter:   "sDC",
		Cluster:      "sCluster2",
		ResourcePool: "sResourcePool",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, converter.AZ{
		Datacenter:   "tDC",
		Cluster:      "tCluster1",
		ResourcePool: "tResourcePool",
		Name:         "AZ1",
	}, result[0])
	require.Equal(t, converter.AZ{
		Datacenter:   "tDC",
		Cluster:      "tCluster2",
		ResourcePool: "tResourcePool",
		Name:         "AZ1",
	}, result[1])
}

func TestAZMappingEquals(t *testing.T) {
	a := converter.AZ{
		Datacenter:   "DC",
		Cluster:      "Cluster",
		ResourcePool: "RP",
		Name:         "AZ1",
	}
	b := converter.AZ{
		Datacenter:   "DC",
		Cluster:      "Cluster",
		ResourcePool: "RP",
		Name:         "AZ1",
	}
	require.True(t, a.Equals(b))

	b.Datacenter = strings.ToLower(b.Datacenter)
	require.True(t, a.Equals(b))

	b.Datacenter = "DC2"
	require.False(t, a.Equals(b))

	b.Datacenter = "DC"
	b.Cluster = "Cluster2"
	require.False(t, a.Equals(b))

	b.Cluster = "Cluster"
	b.ResourcePool = "RP2"
	require.False(t, a.Equals(b))

	b.ResourcePool = "RP"
	b.Name = "AZ2"
	require.False(t, a.Equals(b))
}
