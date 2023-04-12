/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package converter

import (
	"fmt"
	"strings"
)

func TargetFolder(sourceVMFolder string, targetDatacenter string) (string, error) {
	splitFn := func(c rune) bool {
		return c == '/'
	}

	p := strings.FieldsFunc(sourceVMFolder, splitFn)
	if len(p) < 2 {
		return "", fmt.Errorf("expected a source VM folder path of at least 2 parts, but got '%s'", sourceVMFolder)
	}
	if p[1] != "vm" {
		return "", fmt.Errorf("expected a source VM folder path to contain 'vm' in path under datacenter, but got '%s'", p[1])
	}

	p[0] = targetDatacenter
	targetFolder := "/" + strings.Join(p, "/")
	return targetFolder, nil
}
