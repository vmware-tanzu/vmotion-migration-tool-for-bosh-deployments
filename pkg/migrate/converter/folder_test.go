/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter_test

import (
	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/migrate/converter"
	"testing"
)

var testTargetFolderTests = []struct {
	name         string
	inFolder     string
	inDatacenter string
	outPath      string
	err          string
}{
	{
		"VM in different datacenter",
		"/sDC/vm",
		"tDC",
		"/tDC/vm",
		"",
	},
	{
		"VM in same datacenter",
		"/sDC/vm",
		"sDC",
		"/sDC/vm",
		"",
	},
	{
		"VM in sub-folder",
		"/sDC/vm/guid/path",
		"tDC",
		"/tDC/vm/guid/path",
		"",
	},
	{
		"VM with missing path",
		"",
		"tDC",
		"",
		"expected a source VM folder path of at least 2 parts, but got ''",
	},
	{
		"VM with no vm sub-path",
		"/sDC/guid/path",
		"tDC",
		"",
		"expected a source VM folder path to contain 'vm' in path under datacenter, but got 'guid'",
	},
}

func TestTargetFolder(t *testing.T) {
	for _, tt := range testTargetFolderTests {
		t.Run(tt.name, func(t *testing.T) {
			outPath, err := converter.TargetFolder(tt.inFolder, tt.inDatacenter)
			if tt.err != "" {
				require.Error(t, err)
				require.Equal(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.outPath, outPath)
			}
		})
	}
}
