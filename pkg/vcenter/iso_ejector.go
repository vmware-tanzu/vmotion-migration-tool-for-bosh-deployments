/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type ISOEjector struct {
	vm *object.VirtualMachine
}

func NewISOEjector(vm *object.VirtualMachine) *ISOEjector {
	return &ISOEjector{
		vm: vm,
	}
}

func (e *ISOEjector) EjectISO(ctx context.Context) error {
	l := log.FromContext(ctx)
	l.Debugf("Listing %s devicesToChange", e.vm.Name())
	virtualDeviceList, err := e.vm.Device(ctx)
	if err != nil {
		return fmt.Errorf("failed to list devicesToChange for VM %s: %w", e.vm.Name(), err)
	}

	l.Debug("Ejecting env.iso from CD-ROM")
	cd, err := virtualDeviceList.FindCdrom("")
	if err != nil && !strings.Contains(err.Error(), "no cdrom device found") {
		return fmt.Errorf("could not get CD-ROM device on %s: %w", e.vm.Name(), err)
	}
	if cd != nil {
		debugLogCDDrive(l, *cd)
		err = virtualDeviceList.Disconnect(cd)
		if err != nil {
			return fmt.Errorf("could not disconnect CD-ROM from %s: %w", e.vm.Name(), err)
		}
		err = e.vm.EditDevice(ctx, virtualDeviceList.EjectIso(cd))
		if err != nil {
			return fmt.Errorf("could not eject env.iso from %s: %w", e.vm.Name(), err)
		}
	} else {
		l.Debug("Could not find a CD-ROM device, skipping eject")
	}
	return nil
}

func debugLogCDDrive(l *logrus.Entry, cd types.VirtualCdrom) {
	j, err := json.MarshalIndent(cd, "", "  ")
	if err != nil {
		l.Errorf("Could not serialize CD-ROM: %s", err.Error())
	}
	l.Debugln("CD-ROM:")
	l.Debugln(string(j))

}
