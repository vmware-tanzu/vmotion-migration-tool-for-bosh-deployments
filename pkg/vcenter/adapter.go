package vcenter

import (
	"github.com/vmware/govmomi/vim25/types"
)

type anyAdapter struct {
	*types.VirtualVmxnet3
	*types.VirtualE1000
}

func (a *anyAdapter) Device() types.BaseVirtualDevice {
	if a.VirtualE1000 != nil {
		return a.VirtualE1000
	} else if a.VirtualVmxnet3 != nil {
		return a.VirtualVmxnet3
	} else {
		panic("bug: no supported vNICs found")
	}
}
