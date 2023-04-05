/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter_test

import (
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/vcenter"
	"testing"
)

func TestPoolWithSingleVCenter(t *testing.T) {
	p := vcenter.NewPool()
	p.AddSource("az1", "vc01.example.com", "admin", "secret", "dc", true)
	p.AddSource("az2", "vc01.example.com", "admin", "secret", "dc", true)
	p.AddTarget("az1", "vc01.example.com", "admin", "secret", "dc", true)
	p.AddTarget("az2", "vc01.example.com", "admin", "secret", "dc", true)

	sc := p.GetClients()
	require.Len(t, sc, 1)

	c1 := p.GetSourceClientByAZ("az1")
	c2 := p.GetSourceClientByAZ("az2")
	c3 := p.GetTargetClientByAZ("az1")
	c4 := p.GetTargetClientByAZ("az2")
	require.Same(t, c1, c2)
	require.Same(t, c1, c3)
	require.Same(t, c3, c4)
}

func TestPoolWithSeparateVCenterForSourceAndTarget(t *testing.T) {
	p := vcenter.NewPool()
	p.AddSource("az1", "vc01.example.com", "admin", "secret", "dc", true)
	p.AddSource("az2", "vc01.example.com", "admin", "secret", "dc", true)
	p.AddTarget("az1", "vc02.example.com", "admin", "secret", "dc", true)
	p.AddTarget("az2", "vc02.example.com", "admin", "secret", "dc", true)

	sc := p.GetClients()
	require.Len(t, sc, 2)

	c1 := p.GetSourceClientByAZ("az1")
	c2 := p.GetSourceClientByAZ("az2")
	c3 := p.GetTargetClientByAZ("az1")
	c4 := p.GetTargetClientByAZ("az2")
	require.Same(t, c1, c2)
	require.NotSame(t, c1, c3)
	require.Same(t, c3, c4)
}

func TestPoolWithSeparateVCenterForEachAZ(t *testing.T) {
	p := vcenter.NewPool()
	p.AddSource("az1", "vc01.example.com", "admin", "secret", "dc", true)
	p.AddSource("az2", "vc02.example.com", "admin", "secret", "dc", true)
	p.AddTarget("az1", "vc03.example.com", "admin", "secret", "dc", true)
	p.AddTarget("az2", "vc04.example.com", "admin", "secret", "dc", true)

	sc := p.GetClients()
	require.Len(t, sc, 4)

	c1 := p.GetSourceClientByAZ("az1")
	c2 := p.GetSourceClientByAZ("az2")
	c3 := p.GetTargetClientByAZ("az1")
	c4 := p.GetTargetClientByAZ("az2")
	require.NotSame(t, c1, c2)
	require.NotSame(t, c2, c3)
	require.NotSame(t, c3, c4)
}

func TestPoolWithSeparateVCenterAccountsForEachAZ(t *testing.T) {
	p := vcenter.NewPool()
	p.AddSource("az1", "vc01.example.com", "admin1", "secret", "dc", true)
	p.AddSource("az2", "vc01.example.com", "admin2", "secret", "dc", true)
	p.AddTarget("az1", "vc01.example.com", "admin3", "secret", "dc", true)
	p.AddTarget("az2", "vc01.example.com", "admin4", "secret", "dc", true)

	sc := p.GetClients()
	require.Len(t, sc, 4)

	c1 := p.GetSourceClientByAZ("az1")
	c2 := p.GetSourceClientByAZ("az2")
	c3 := p.GetTargetClientByAZ("az1")
	c4 := p.GetTargetClientByAZ("az2")
	require.NotSame(t, c1, c2)
	require.NotSame(t, c2, c3)
	require.NotSame(t, c3, c4)
}

func TestAddingSameAZTwice(t *testing.T) {
	p := vcenter.NewPool()
	p.AddSource("az1", "vc01.example.com", "admin1", "secret", "dc", true)
	p.AddSource("az1", "vc01.example.com", "admin1", "secret", "dc", true)

	sc := p.GetClients()
	require.Len(t, sc, 1)

	c1 := p.GetSourceClientByAZ("az1")
	require.NotNil(t, c1)
}

func TestGetNonExistentClient(t *testing.T) {
	p := vcenter.NewPool()
	p.AddSource("az1", "vc01.example.com", "admin1", "secret", "dc", true)

	c := p.GetSourceClientByAZ("az2")
	require.Nil(t, c)
	c = p.GetTargetClientByAZ("az2")
	require.Nil(t, c)
}
