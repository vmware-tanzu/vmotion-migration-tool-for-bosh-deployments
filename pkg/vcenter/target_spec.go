/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

type TargetSpec struct {
	Name         string
	Datacenter   string
	Cluster      string
	ResourcePool string
	Folder       string
	Datastores   map[string]string
	Networks     map[string]string
}
