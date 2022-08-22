/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package duration

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func HumanReadable(d time.Duration) string {
	hours := int64(d.Hours())
	minutes := int64(math.Mod(d.Minutes(), 60))
	seconds := int64(math.Mod(d.Seconds(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"h", hours},
		{"m", minutes},
		{"s", seconds},
	}

	var parts []string
	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		default:
			parts = append(parts, fmt.Sprintf("%d%s", chunk.amount, chunk.singularName))
		}
	}

	return strings.Join(parts, " ")
}
