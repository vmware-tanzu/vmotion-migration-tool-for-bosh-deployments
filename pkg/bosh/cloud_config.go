/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package bosh

type CloudConfig struct {
	AZs []AZs `yaml:"azs"`
}

type AZs struct {
	CPI  string `yaml:"cpi"`
	Name string `yaml:"name"`
}
