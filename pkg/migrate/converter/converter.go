package converter

import (
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

type NetworkMapper interface {
	TargetNetworks(sourceVM *vcenter.VM) (map[string]string, error)
}

type ResourcePoolMapper interface {
	TargetResourcePool(sourceVM *vcenter.VM) (string, error)
}

type Converter struct {
	rpMapper         ResourcePoolMapper
	netMapper        NetworkMapper
	targetDatacenter string
	targetCluster    string
	targetDatastore  string
}

func New(net NetworkMapper, rp ResourcePoolMapper, targetDatacenter, targetCluster, targetDatastore string) *Converter {
	return &Converter{
		rpMapper:         rp,
		netMapper:        net,
		targetDatacenter: targetDatacenter,
		targetCluster:    targetCluster,
		targetDatastore:  targetDatastore,
	}
}

func (c *Converter) TargetSpec(sourceVM *vcenter.VM) (*vcenter.TargetSpec, error) {
	rp, err := c.rpMapper.TargetResourcePool(sourceVM)
	if err != nil {
		return nil, err
	}
	nets, err := c.netMapper.TargetNetworks(sourceVM)
	if err != nil {
		return nil, err
	}

	return &vcenter.TargetSpec{
		Name:         sourceVM.Name,
		Datacenter:   c.targetDatacenter,
		Cluster:      c.targetCluster,
		ResourcePool: rp,
		Datastore:    c.targetDatastore,
		Networks:     nets,
	}, nil
}
