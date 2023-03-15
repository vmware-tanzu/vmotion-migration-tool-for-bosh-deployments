/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import (
	"github.com/stretchr/testify/require"
	"github.com/vmware/govmomi/vim25/types"
	"testing"
)

func TestVirtualEthernetCardNetworkBackingInfo(t *testing.T) {
	a := anyNetworkBackingInfo{
		info: &types.VirtualEthernetCardNetworkBackingInfo{
			Network: &types.ManagedObjectReference{
				Value: "network-25",
			},
		},
	}
	require.Equal(t, "network-25", a.NetworkID())
}
