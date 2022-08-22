/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate/converter"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
)

var explicitTests = []struct {
	name string
	in   *vcenter.VM
	out  *vcenter.TargetSpec
}{
	{
		"Standard VM",
		&vcenter.VM{
			Name:         "virtualMachine42",
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Datastore:    "sDS",
			Networks:     []string{"sN"},
		},
		&vcenter.TargetSpec{
			Name:         "virtualMachine42",
			Datacenter:   "tDC",
			Cluster:      "tC",
			ResourcePool: "tRP",
			Datastore:    "tDS",
			Networks:     map[string]string{"sN": "tN"},
		},
	},
}

func TestExplicitConverter(t *testing.T) {
	rp := converter.NewExplicitResourcePool("tRP")
	net := converter.NewEmptyMappedNetwork().Add("sN", "tN")
	c := converter.New(net, rp, "tDC", "tC", "tDS")
	for _, tt := range explicitTests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := c.TargetSpec(tt.in)
			require.NoError(t, err)
			require.Equal(t, tt.out, spec)
		})
	}
}

var mappedTests = []struct {
	name string
	in   *vcenter.VM
	out  *vcenter.TargetSpec
	err  string
}{
	{
		"Standard VM",
		&vcenter.VM{
			Name:         "virtualMachine42",
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Datastore:    "sDS",
			Networks:     []string{"sN2"},
		},
		&vcenter.TargetSpec{
			Name:         "virtualMachine42",
			Datacenter:   "tDC",
			Cluster:      "tC",
			ResourcePool: "tRP",
			Datastore:    "tDS",
			Networks:     map[string]string{"sN2": "tN2"},
		}, "",
	},
	{
		"Unmapped Network",
		&vcenter.VM{
			Name:         "VM3",
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Datastore:    "sDS",
			Networks:     []string{"sN-missing"},
		},
		&vcenter.TargetSpec{},
		"could not find a target network for VM VM3 attached to network sN-missing: ensure you add a corresponding network mapping to the config file",
	},
	{
		"Unmapped Resource Pool",
		&vcenter.VM{
			Name:         "VM12",
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP-missing",
			Datastore:    "sDS",
			Networks:     []string{"sN"},
		},
		&vcenter.TargetSpec{},
		"could not find a target resource pool for VM VM12 in resource pool sRP-missing: ensure you add a corresponding resource pool mapping to the config file",
	},
	{
		"Stemcell (no network)",
		&vcenter.VM{
			Name:         "sc-someguid",
			Datacenter:   "sDC",
			Cluster:      "sC",
			ResourcePool: "sRP",
			Datastore:    "sDS",
		},
		&vcenter.TargetSpec{
			Name:         "sc-someguid",
			Datacenter:   "tDC",
			Cluster:      "tC",
			ResourcePool: "tRP",
			Datastore:    "tDS",
			Networks:     map[string]string{},
		}, "",
	},
}

func TestMappedConverter(t *testing.T) {
	rp := converter.NewMappedResourcePool(map[string]string{
		"sRP":  "tRP",
		"sRP2": "tRP2",
	})
	net := converter.NewMappedNetwork(map[string]string{
		"sN":  "tN",
		"sN2": "tN2",
		"sN3": "tN3",
	})
	c := converter.New(net, rp, "tDC", "tC", "tDS")
	for _, tt := range mappedTests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := c.TargetSpec(tt.in)
			if tt.err != "" {
				require.Error(t, err)
				require.Equal(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.out, spec)
			}
		})
	}
}
