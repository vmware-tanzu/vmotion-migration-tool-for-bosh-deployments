package vcenter_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

func TestFullyQualifiedResourcePool(t *testing.T) {
	spec := vcenter.TargetSpec{
		Datacenter:   "dc",
		Cluster:      "cl",
		ResourcePool: "rp",
	}
	require.Equal(t, "/dc/host/cl/Resources/rp", spec.FullyQualifiedResourcePool())
}

func TestFullyQualifiedResourcePoolWithoutRP(t *testing.T) {
	spec := vcenter.TargetSpec{
		Datacenter: "dc",
		Cluster:    "cl",
	}
	require.Equal(t, "/dc/host/cl/Resources", spec.FullyQualifiedResourcePool())
}
