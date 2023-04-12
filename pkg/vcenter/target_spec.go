/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import (
	"fmt"
	"strings"

	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
)

type TargetSpec struct {
	Name         string
	Datacenter   string
	Cluster      string
	ResourcePool string
	Folder       string
	Datastores   map[string]string
	Networks     map[string]string
}

// FullyQualifiedResourcePool ensures we avoid "multiple found" errors
func (t TargetSpec) FullyQualifiedResourcePool() string {
	rp := t.ResourcePool
	if !strings.Contains(rp, "/") {
		rp = fmt.Sprintf("/%s/host/%s/Resources", t.Datacenter, t.Cluster)
		if t.ResourcePool != "" {
			rp = rp + "/" + t.ResourcePool
		}
		log.WithoutContext().Debugf("Found short resource pool name adjusting to %s", rp)
	}
	return rp
}
