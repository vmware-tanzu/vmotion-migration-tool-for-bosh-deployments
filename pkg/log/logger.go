/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package log

import (
	"context"
	"github.com/sirupsen/logrus"
)

type TaskIDKeyType int

const TaskIDKey TaskIDKeyType = 0

func Initialize(debug bool) {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.ErrorLevel)
	}
}

func WithoutContext() *logrus.Entry {
	return logrus.NewEntry(logrus.StandardLogger())
}

func FromContext(ctx context.Context) *logrus.Entry {
	taskID := taskIDFromContext(ctx)
	if taskID == -1 {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.WithFields(logrus.Fields{"TaskID": taskID})
}

func taskIDFromContext(ctx context.Context) int {
	id, ok := ctx.Value(TaskIDKey).(int)
	if ok {
		return id
	}
	return -1
}
