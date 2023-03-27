/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package log

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
)

type BufferedStdout struct {
	sb strings.Builder
}

func NewBufferedStdout() *BufferedStdout {
	return &BufferedStdout{}
}

// Println outputs an empty line
func (p *BufferedStdout) Println() {
	p.Print("")
}

// Print outputs the specified string as-is on the current line
func (p *BufferedStdout) Print(s string) {
	id := uuid.New()
	p.PrintUpdatable(id.String(), s)
}

// Printf formats and outputs the format string on the current line
func (p *BufferedStdout) Printf(format string, a ...interface{}) {
	p.Print(fmt.Sprintf(format, a...))
}

// PrintUpdatablef formats and outputs the format string on the current line, or updates it in place
func (p *BufferedStdout) PrintUpdatablef(id, format string, a ...interface{}) {
	p.PrintUpdatable(id, fmt.Sprintf(format, a...))
}

// PrintUpdatable outputs the specified string on the current line, or updates it in place
func (p *BufferedStdout) PrintUpdatable(id, s string) {
	p.sb.WriteString(s)
}

func (p *BufferedStdout) String() string {
	return p.sb.String()
}
