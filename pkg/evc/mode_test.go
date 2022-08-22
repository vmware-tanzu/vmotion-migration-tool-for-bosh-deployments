package evc_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/bosh"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/evc/evcfakes"
	"testing"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/evc"
)

func TestSetEVCMode(t *testing.T) {
	instances := []bosh.Instance{
		{
			VMName:     "vm-fb00972b-9b3d-4ca8-a988-97768e681135",
			Name:       "compute/803fd123-386b-4caa-aecd-88c24d0eac72",
			Job:        "compute",
			Deployment: "cf-7e5204ac0b19b238008e",
		},
		{
			VMName:     "vm-f1c365b4-2582-4a72-ae0f-4d8ab52c5ba3",
			Name:       "compute/0835d357-3eac-4c87-8ee6-de6c0d0fbdf2",
			Job:        "compute",
			Deployment: "cf-7e5204ac0b19b238008e",
		},
		{
			VMName:     "vm-cf0c324c-1b3b-4abe-b4be-289bdc4c1b86",
			Name:       "router/296699c8-c39b-41b6-8fdd-1a23972bd1b7",
			Job:        "router",
			Deployment: "cf-7e5204ac0b19b238008e",
		},
	}

	vc := &evcfakes.FakeVCenterClient{}
	var evcSetVMs = make([]string, 0)
	vc.SetEVCModeCalls(func(ctx context.Context, vmName, evcModeName string) error {
		evcSetVMs = append(evcSetVMs, vmName)
		require.Equal(t, "intel-broadwell", evcModeName)
		return nil
	})

	boshClient := &evcfakes.FakeBoshClient{}
	boshClient.InstancesReturns(instances, nil)

	var stoppedInstances = make([]bosh.Instance, 0)
	boshClient.StopCalls(func(ctx context.Context, i bosh.Instance) error {
		stoppedInstances = append(stoppedInstances, i)
		return nil
	})
	var startedInstances = make([]bosh.Instance, 0)
	boshClient.StartCalls(func(ctx context.Context, i bosh.Instance) error {
		startedInstances = append(startedInstances, i)
		return nil
	})

	m := evc.New(vc, boshClient)
	err := m.Set(context.TODO(), "intel-broadwell")
	require.NoError(t, err)

	require.ElementsMatch(t, instances, stoppedInstances)
	require.ElementsMatch(t, instances, startedInstances)
	require.ElementsMatch(t, []string{
		"vm-fb00972b-9b3d-4ca8-a988-97768e681135",
		"vm-f1c365b4-2582-4a72-ae0f-4d8ab52c5ba3",
		"vm-cf0c324c-1b3b-4abe-b4be-289bdc4c1b86",
	}, evcSetVMs)
}
