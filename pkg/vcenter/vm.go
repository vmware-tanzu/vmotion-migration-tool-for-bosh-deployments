/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

type VM struct {
	Name         string
	AZ           string
	Datacenter   string
	Cluster      string
	ResourcePool string
	Disks        []Disk
	Networks     []string
}

type Disk struct {
	ID        int32
	Datastore string
}

type VMNotFoundError struct {
	Name string
}

func NewVMNotFoundError(name string) error {
	return &VMNotFoundError{
		Name: name,
	}
}

func (e *VMNotFoundError) Error() string {
	return "VM not found: " + e.Name
}
