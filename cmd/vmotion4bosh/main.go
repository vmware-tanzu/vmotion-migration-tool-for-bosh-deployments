/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"os"

	goflags "github.com/jessevdk/go-flags"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/command"
)

type CommandHolder struct {
	Version             command.VersionCommand `command:"version" description:"Print version information and exit"`
	CrossVCenterMigrate command.MigrateVM      `command:"migrate-vm" description:"Migrates a vm from one vcenter to another"`
	Migrate             command.Migrate        `command:"migrate" description:"Migrates an entire foundation from one vcenter to another"`
	Revert              command.Revert         `command:"revert" description:"Reverts a prior migration back to the source vcenter"`
}

var Command CommandHolder

func main() {
	parser := goflags.NewParser(&Command, goflags.HelpFlag)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
