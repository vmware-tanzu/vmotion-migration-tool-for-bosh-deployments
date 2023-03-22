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
		Cluster:     "sC1",
		VCenterHost: "VC1",
	}, converter.AZMapping{
		Cluster:     "tC1",
		VCenterHost: "VC1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:     "sC1",
		VCenterHost: "VC1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:     "tC1",
		VCenterHost: "VC1",
	}, result)
}

func TestTargetComputeOneToOneClusterDifferentVCenter(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster:     "sC1",
		VCenterHost: "sVC1",
	}, converter.AZMapping{
		Cluster:     "tC1",
		VCenterHost: "tVC1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:     "sC1",
		VCenterHost: "sVC1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:     "tC1",
		VCenterHost: "tVC1",
	}, result)
}

func TestTargetComputeWithDefaultResourcePool(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster:     "sC1",
		VCenterHost: "sVC1",
	}, converter.AZMapping{
		Cluster:     "tC1",
		VCenterHost: "tVC1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:      "sC1",
		VCenterHost:  "sVC1",
		ResourcePool: "Resources",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:     "tC1",
		VCenterHost: "tVC1",
	}, result)
}

func TestTargetComputeOneToOneClusterAndRP(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		VCenterHost:  "sVC1",
	}, converter.AZMapping{
		Cluster:      "tC1",
		ResourcePool: "tRP1",
		VCenterHost:  "tVC1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		VCenterHost:  "sVC1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:      "tC1",
		ResourcePool: "tRP1",
		VCenterHost:  "tVC1",
	}, result)
}

func TestTargetComputeOneToOneClusterAndOneSrcRP(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		VCenterHost:  "sVC1",
	}, converter.AZMapping{
		Cluster:     "tC1",
		VCenterHost: "tVC1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		VCenterHost:  "sVC1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:     "tC1",
		VCenterHost: "tVC1",
	}, result)
}

func TestTargetComputeOneToOneClusterAndOneTargetRP(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster:     "sC1",
		VCenterHost: "sVC1",
	}, converter.AZMapping{
		Cluster:      "tC1",
		ResourcePool: "tRP1",
		VCenterHost:  "tVC1",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:     "sC1",
		VCenterHost: "sVC1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:      "tC1",
		ResourcePool: "tRP1",
		VCenterHost:  "tVC1",
	}, result)
}

func TestTargetOneClusterAndRPsToThreeClustersOnDifferentVCenters(t *testing.T) {
	cm := converter.NewEmptyMappedCompute()
	cm.Add(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		VCenterHost:  "sVC1",
	}, converter.AZMapping{
		Cluster:      "tC1",
		ResourcePool: "ResourcePool",
		VCenterHost:  "tVC1",
	})
	cm.Add(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP2",
		VCenterHost:  "sVC1",
	}, converter.AZMapping{
		Cluster:      "tC2",
		ResourcePool: "ResourcePool",
		VCenterHost:  "tVC2",
	})
	cm.Add(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP3",
		VCenterHost:  "sVC1",
	}, converter.AZMapping{
		Cluster:      "tC3",
		ResourcePool: "ResourcePool",
		VCenterHost:  "tVC3",
	})

	result, err := cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP1",
		VCenterHost:  "sVC1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:      "tC1",
		ResourcePool: "ResourcePool",
		VCenterHost:  "tVC1",
	}, result)

	result, err = cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP2",
		VCenterHost:  "sVC1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:      "tC2",
		ResourcePool: "ResourcePool",
		VCenterHost:  "tVC2",
	}, result)

	result, err = cm.TargetComputeFromSource(converter.AZMapping{
		Cluster:      "sC1",
		ResourcePool: "sRP3",
		VCenterHost:  "sVC1",
	})
	require.NoError(t, err)
	require.Equal(t, converter.AZMapping{
		Cluster:      "tC3",
		ResourcePool: "ResourcePool",
		VCenterHost:  "tVC3",
	}, result)
}
