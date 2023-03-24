/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter_test

import (
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate/converter"
	"testing"
)

func TestTargetComputeOneToOneClusterSameVCenter(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	}, converter.AZMapping{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	}, result)
}

func TestTargetComputeOneToOneClusterDifferentVCenter(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	}, converter.AZMapping{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	}, result)
}

func TestTargetComputeWithDefaultResourcePool(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	}, converter.AZMapping{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		Name:         "AZ1",
		ResourcePool: "Resources",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	}, result)
}

func TestTargetComputeOneToOneClusterAndRP(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		Name:         "AZ1",
	}, converter.AZMapping{
		Datacenter:   "tDC1",
		Cluster:      "tC1",
		ResourcePool: "tRP1",
		Name:         "AZ1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Datacenter:   "tDC1",
		Cluster:      "tC1",
		ResourcePool: "tRP1",
		Name:         "AZ1",
	}, result)
}

func TestTargetComputeOneToOneClusterAndOneSrcRP(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		Name:         "AZ1",
	}, converter.AZMapping{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Datacenter: "tDC1",
		Cluster:    "tC1",
		Name:       "AZ1",
	}, result)
}

func TestTargetComputeOneToOneClusterAndOneTargetRP(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	}, converter.AZMapping{
		Datacenter:   "tDC1",
		Cluster:      "tC1",
		ResourcePool: "tRP1",
		Name:         "AZ1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Datacenter: "sDC1",
		Cluster:    "sC1",
		Name:       "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Datacenter:   "tDC1",
		Cluster:      "tC1",
		ResourcePool: "tRP1",
		Name:         "AZ1",
	}, result)
}

func TestTargetOneClusterAndRPsToThreeClustersOnDifferentVCenters(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		Name:         "AZ1",
	}, converter.AZMapping{
		Datacenter:   "tDC1",
		Cluster:      "tC1",
		ResourcePool: "ResourcePool",
		Name:         "AZ1",
	})
	cm.Add(converter.AZMapping{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP2",
		Name:         "AZ1",
	}, converter.AZMapping{
		Datacenter:   "tDC1",
		Cluster:      "tC2",
		ResourcePool: "ResourcePool",
		Name:         "AZ2",
	})
	cm.Add(converter.AZMapping{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP3",
		Name:         "AZ1",
	}, converter.AZMapping{
		Datacenter:   "tDC1",
		Cluster:      "tC3",
		ResourcePool: "ResourcePool",
		Name:         "AZ3",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Datacenter:   "tDC1",
		Cluster:      "tC1",
		ResourcePool: "ResourcePool",
		Name:         "AZ1",
	}, result)

	result, err = cm.TargetComputeFromSource(converter.AZMapping{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP2",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Datacenter:   "tDC1",
		Cluster:      "tC2",
		ResourcePool: "ResourcePool",
		Name:         "AZ2",
	}, result)

	result, err = cm.TargetComputeFromSource(converter.AZMapping{
		Datacenter:   "sDC1",
		Cluster:      "sC1",
		ResourcePool: "sRP3",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Datacenter:   "tDC1",
		Cluster:      "tC3",
		ResourcePool: "ResourcePool",
		Name:         "AZ3",
	}, result)
}

func TestTargetManyToManyClustersInAZ(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Datacenter:   "sDC",
		Cluster:      "sCluster1",
		ResourcePool: "sResourcePool",
		Name:         "AZ1",
	}, converter.AZMapping{
		Datacenter:   "tDC",
		Cluster:      "tCluster1",
		ResourcePool: "tResourcePool",
		Name:         "AZ1",
	})
	cm.Add(converter.AZMapping{
		Datacenter:   "sDC",
		Cluster:      "sCluster2",
		ResourcePool: "sResourcePool",
		Name:         "AZ1",
	}, converter.AZMapping{
		Datacenter:   "tDC",
		Cluster:      "tCluster2",
		ResourcePool: "tResourcePool",
		Name:         "AZ1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Datacenter:   "sDC",
		Cluster:      "sCluster1",
		ResourcePool: "sResourcePool",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Datacenter:   "tDC",
		Cluster:      "tCluster1",
		ResourcePool: "tResourcePool",
		Name:         "AZ1",
	}, result)
	result, err = cm.TargetComputeFromSource(converter.AZMapping{
		Datacenter:   "sDC",
		Cluster:      "sCluster2",
		ResourcePool: "sResourcePool",
		Name:         "AZ1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Datacenter:   "tDC",
		Cluster:      "tCluster2",
		ResourcePool: "tResourcePool",
		Name:         "AZ1",
	}, result)
}
