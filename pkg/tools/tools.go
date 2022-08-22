//go:build tools

/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package tools

import (
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
)

// This file imports packages that are used when running go generate, or used
// during the development process but not otherwise depended on by built code.
