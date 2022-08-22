/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package thumbprint_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/thumbprint"
)

func TestRetrieveSHA1InvalidHost(t *testing.T) {
	tp, err := thumbprint.RetrieveSHA1("www.invalid.tld", 443)

	require.Error(t, err)
	require.Empty(t, tp)
}

func TestRetrieveSHA1(t *testing.T) {
	tp, err := thumbprint.RetrieveSHA1("www.vmware.com", 443)
	require.NoError(t, err)
	require.NotEmpty(t, tp)
	fmt.Println(tp)
}
