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

func TestTargetComputeOneToOneCluster(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster: "sC1",
	}, converter.AZMapping{
		Cluster: "tC1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster: "sC1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster: "tC1",
	}, result)
}

func TestTargetComputeWithDefaultResourcePool(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster: "sC1",
	}, converter.AZMapping{
		Cluster: "tC1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "Resources",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster: "tC1",
	}, result)
}

func TestTargetComputeOneToOneClusterAndRP(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP1",
	}, converter.AZMapping{
		Cluster:      "tC1",
		ResourcePool: "tRP1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:      "tC1",
		ResourcePool: "tRP1",
	}, result)
}

func TestTargetComputeOneToOneClusterAndOneSrcRP(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP1",
	}, converter.AZMapping{
		Cluster: "tC1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster: "tC1",
	}, result)
}

func TestTargetComputeOneToOneClusterAndOneTargetRP(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster: "sC1",
	}, converter.AZMapping{
		Cluster:      "tC1",
		ResourcePool: "tRP1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster: "sC1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:      "tC1",
		ResourcePool: "tRP1",
	}, result)
}

func TestTargetOneClusterAndRPsToThreeClusters(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP1",
	}, converter.AZMapping{
		Cluster:      "tC1",
		ResourcePool: "ResourcePool",
	})
	cm.Add(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP2",
	}, converter.AZMapping{
		Cluster:      "tC2",
		ResourcePool: "ResourcePool",
	})
	cm.Add(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP3",
	}, converter.AZMapping{
		Cluster:      "tC3",
		ResourcePool: "ResourcePool",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:      "tC1",
		ResourcePool: "ResourcePool",
	}, result)

	result, err = cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP2",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:      "tC2",
		ResourcePool: "ResourcePool",
	}, result)

	result, err = cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP3",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:      "tC3",
		ResourcePool: "ResourcePool",
	}, result)
}
