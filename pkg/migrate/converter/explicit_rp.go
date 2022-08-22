/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter

import "github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"

type ExplicitRP struct {
	target string
}

func NewExplicitResourcePool(target string) *ExplicitRP {
	return &ExplicitRP{
		target: target,
	}
}

func (c *ExplicitRP) TargetResourcePool(sourceVM *vcenter.VM) (string, error) {
	return c.target, nil
}
