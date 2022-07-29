package vcenter_test

import (
	"context"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

func VPXTest(f func(context.Context, *govmomi.Client)) {
	model := simulator.VPX()
	defer model.Remove()
	model.Pool = 1

	simulator.Test(func(ctx context.Context, vimClient *vim25.Client) {
		c := &govmomi.Client{
			Client:         vimClient,
			SessionManager: session.NewManager(vimClient),
		}
		f(ctx, c)
	}, model)
}

func findSimulatorObject(kind, name string) mo.Entity {
	for _, o := range simulator.Map.All(kind) {
		if o.Entity().Name == name {
			return o
		}
	}
	return nil
}
