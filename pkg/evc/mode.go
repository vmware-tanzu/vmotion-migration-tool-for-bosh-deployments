package evc

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/bosh"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/config"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/worker"
)

//counterfeiter:generate . BoshClient
type BoshClient interface {
	Instances(context.Context) ([]bosh.Instance, error)
	Stop(ctx context.Context, instance bosh.Instance) error
	Start(ctx context.Context, instance bosh.Instance) error
}

//counterfeiter:generate . VCenterClient
type VCenterClient interface {
	SetEVCMode(ctx context.Context, vmName, evcModeName string) error
}

type Mode struct {
	sourceVCenter VCenterClient
	sourceBosh    BoshClient
}

func New(sourceVCenter VCenterClient, sourceBosh BoshClient) *Mode {
	return &Mode{
		sourceVCenter: sourceVCenter,
		sourceBosh:    sourceBosh,
	}
}

func SetOnAllSourceVMs(evcModeName string, c config.Config, ctx context.Context) error {
	//sourceVCenter := vcenter.New(
	//	c.Source.VCenter.Host, c.Source.VCenter.Username, c.Source.VCenter.Password, c.Source.VCenter.Insecure)
	//defer sourceVCenter.Logout(ctx)
	//
	//sourceBosh := bosh.New(c.Source.Bosh.Host, c.Source.Bosh.ClientID, c.Source.Bosh.ClientSecret)
	//
	//m := New(sourceVCenter, sourceBosh)
	//return m.Set(ctx, evcModeName)
	return nil
}

type evcResult struct {
	instanceName string
	err          error
}

func (er evcResult) Success() bool {
	return er.err == nil
}

func (m *Mode) Set(ctx context.Context, evcModeName string) error {
	l := log.FromContext(ctx)

	instances, err := m.sourceBosh.Instances(ctx)
	if err != nil {
		return err
	}

	// sort all instances into buckets by job type
	instancesByJob := make(map[string][]bosh.Instance, len(instances))
	for _, i := range instances {
		if _, ok := instancesByJob[i.Job]; !ok {
			instancesByJob[i.Job] = make([]bosh.Instance, 0)
		}
		instancesByJob[i.Job] = append(instancesByJob[i.Job], i)
	}

	// create a worker for each job type
	workers := worker.NewPool(len(instancesByJob))
	workers.Start(ctx)
	results := make(chan evcResult, len(instances))

	// set EVC modes in parallel, but only do one VM at a time per job type
	for _, i := range instancesByJob {
		jobInstances := i
		workers.AddTask(func(taskCtx context.Context) {
			for _, ji := range jobInstances {
				l.Debugf("Setting EVC mode %s on %s (%s)", evcModeName, ji.Name, ji.VMName)

				jiErr := m.sourceBosh.Stop(ctx, ji)
				if jiErr == nil {
					jiErr = m.sourceVCenter.SetEVCMode(ctx, ji.VMName, evcModeName)
					if jiErr == nil {
						jiErr = m.sourceBosh.Start(ctx, ji)
					}
				}

				results <- evcResult{
					instanceName: ji.Name,
					err:          jiErr,
				}
			}
		})
	}

	failCount := 0
	for i := 0; i < len(instances); i++ {
		res := <-results
		if !res.Success() {
			failCount++
			l.Warnf("%s - failed setting EVC mode: %s", res.instanceName, res.err)
		}
	}
	close(results)

	return nil
}
