/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import "fmt"

type VM struct {
	Name         string
	AZ           string
	Datacenter   string
	Cluster      string
	ResourcePool string
	Folder       string
	Disks        []Disk
	Networks     []string
}

type Disk struct {
	ID        int32
	Datastore string
}

type VMNotFoundError struct {
	Name string
	Err  error
}

func NewVMNotFoundError(name string, err error) error {
	return &VMNotFoundError{
		Name: name,
		Err:  err,
	}
}

func (e *VMNotFoundError) Error() string {
	return fmt.Sprintf("%s VM not found: %s", e.Name, e.Err)
}
