package vcenter

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type VMRelocator struct {
	DryRun              bool
	sourceClient        *Client
	destinationClient   *Client
	destinationHostPool *HostPool
	updatableStdout     *log.UpdatableStdout
	dryRunMutex         sync.Mutex
}

func NewVMRelocator(sourceClient *Client, destinationClient *Client, destinationHostPool *HostPool, updatableStdout *log.UpdatableStdout) *VMRelocator {
	return &VMRelocator{
		sourceClient:        sourceClient,
		destinationClient:   destinationClient,
		destinationHostPool: destinationHostPool,
		updatableStdout:     updatableStdout,
	}
}

func (r *VMRelocator) WithDryRun(dryRun bool) *VMRelocator {
	r.DryRun = dryRun
	return r
}

func (r *VMRelocator) RelocateVM(ctx context.Context, srcVM *VM, vmTargetSpec *TargetSpec) error {
	l := log.FromContext(ctx)
	l.Infof("Starting %s migration", srcVM.Name)

	sourceVM, err := r.sourceVM(ctx, srcVM)
	if err != nil {
		return err
	}

	targetHost, err := r.destinationHostPool.WaitForLeaseAvailableHost(ctx, vmTargetSpec.Cluster)
	if err != nil {
		return err
	}
	defer r.destinationHostPool.Release(ctx, targetHost)

	relocateSpecBuilder := NewRelocateSpec(r.sourceClient, r.destinationClient).
		WithTargetSpec(vmTargetSpec).
		WithTargetHost(targetHost).
		WithSourceVM(srcVM)

	spec, err := relocateSpecBuilder.Build(ctx)
	if err != nil {
		return err
	}

	// output what we expect to do, everything after this will mutate state
	debugLogRelocateSpec(l, *spec)
	if r.DryRun {
		r.printDryRunOverview(srcVM, vmTargetSpec)
		return nil
	}

	// eject the CD-ROM to avoid host device missing errors
	ejector := NewISOEjector(sourceVM)
	err = ejector.EjectISO(ctx)
	if err != nil {
		return err
	}

	return r.moveVM(ctx, sourceVM, spec)
}

func (r *VMRelocator) moveVM(ctx context.Context, sourceVM *object.VirtualMachine, spec *types.VirtualMachineRelocateSpec) error {
	task, err := sourceVM.Relocate(ctx, *spec, types.VirtualMachineMovePriorityHighPriority)
	if err != nil {
		return fmt.Errorf("failed to migrate %s: %w", sourceVM.Name(), err)
	}

	progressLogger := NewProgressLogger(r.updatableStdout)
	progressSink := progressLogger.NewProgressSink(sourceVM.Name())
	_, err = task.WaitForResult(ctx, progressSink)
	if err != nil {
		return err
	}

	return nil
}

func (r *VMRelocator) sourceVM(ctx context.Context, srcVM *VM) (*object.VirtualMachine, error) {
	srcClient, err := r.sourceClient.getOrCreateUnderlyingClient(ctx)
	if err != nil {
		return nil, err
	}
	f := NewFinder(srcVM.Datacenter, srcClient)
	return f.VirtualMachine(ctx, srcVM.Name)
}

func (r *VMRelocator) printDryRunOverview(srcVM *VM, vmTargetSpec *TargetSpec) {
	// ensure only one VM's details are printed at a time (i.e. whole across multiple lines)
	r.dryRunMutex.Lock()
	defer r.dryRunMutex.Unlock()

	r.updatableStdout.Printf("[dry-run] would migrate %s to:", srcVM.Name)
	r.updatableStdout.Printf("  vcenter:       %s", r.destinationClient.Host)
	r.updatableStdout.Printf("  datacenter:    %s", vmTargetSpec.Datacenter)
	r.updatableStdout.Printf("  cluster:       %s", vmTargetSpec.Cluster)
	r.updatableStdout.Printf("  resource pool: %s", vmTargetSpec.ResourcePool)
	for _, v := range vmTargetSpec.Networks {
		r.updatableStdout.Printf("  network:       %s", v)
	}
	r.updatableStdout.Printf("  datastore:     %s", vmTargetSpec.Datastore)
}

func debugLogRelocateSpec(l *logrus.Entry, spec types.VirtualMachineRelocateSpec) {
	// hide the password while dumping the spec to the log
	c := spec.Service.Credential.(*types.ServiceLocatorNamePassword)
	password := c.Password
	defer func() { c.Password = password }()
	c.Password = "..."

	j, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		l.Errorf("Could not serialize move spec: %s", err.Error())
	}
	l.Debugln("VirtualMachineRelocateSpec:")
	l.Debugln(string(j))
}
