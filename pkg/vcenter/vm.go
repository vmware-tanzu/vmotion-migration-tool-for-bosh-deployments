/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

type VM struct {
	Name         string
	Datacenter   string
	Cluster      string
	ResourcePool string
	Datastore    string
	Networks     []string
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
