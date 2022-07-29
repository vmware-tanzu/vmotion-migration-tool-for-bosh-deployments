package vcenter_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
	"github.com/vmware/govmomi"
)

func TestLeaseAvailableHost(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		vcenterClient := vcenter.NewFromGovmomiClient(client)

		hostPool := vcenter.NewHostPool(vcenterClient, "DC0")
		hostPool.MaxLeasePerHost = 2
		err := hostPool.Initialize(ctx)
		require.NoError(t, err)

		// lease a host and release it
		host, err := hostPool.LeaseAvailableHost(ctx, "DC0_C0")
		require.NoError(t, err)
		require.NotNil(t, host)
		require.Contains(t, host.Name(), "DC0_C0_H")
		hostPool.Release(ctx, host)

		// lease all 3 hosts twice
		for i := 0; i < 6; i++ {
			host, err = hostPool.LeaseAvailableHost(ctx, "DC0_C0")
			require.NoError(t, err)
			require.NotNil(t, host)
			require.Contains(t, host.Name(), "DC0_C0_H")
		}

		// try to get another lease, should fail and return nil
		host, err = hostPool.LeaseAvailableHost(ctx, "DC0_C0")
		require.NoError(t, err)
		require.Nil(t, host)
	})
}

func TestLeaseAvailableHostMin(t *testing.T) {
	VPXTest(func(ctx context.Context, client *govmomi.Client) {
		vcenterClient := vcenter.NewFromGovmomiClient(client)

		hostPool := vcenter.NewHostPool(vcenterClient, "DC0")
		hostPool.MaxLeasePerHost = 1
		err := hostPool.Initialize(ctx)
		require.NoError(t, err)

		// lease all 3 hosts once
		for i := 0; i < 3; i++ {
			host, err := hostPool.LeaseAvailableHost(ctx, "DC0_C0")
			require.NoError(t, err)
			require.NotNil(t, host)
			require.Contains(t, host.Name(), "DC0_C0_H")
		}

		// try to get another lease, should fail and return nil
		host, err := hostPool.LeaseAvailableHost(ctx, "DC0_C0")
		require.NoError(t, err)
		require.Nil(t, host)
	})
}
