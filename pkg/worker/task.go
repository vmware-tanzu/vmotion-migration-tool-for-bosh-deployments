/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package worker

import "context"

type TaskFn func(context.Context)

type task struct {
	fn TaskFn
	id int
}
