/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

// Execute runs root cmd.
func Execute(ctx context.Context, args []string) {
	if err := newCmd().RunContext(ctx, args); err != nil {
		fmt.Fprintf(os.Stdout, "%v\n\n", err)
		os.Exit(1)
	}
}

const unknown = "unknown"

// variables are expected to be set at build time.
var (
	releaseVersion = unknown
	releaseCommit  = unknown
	releaseOS      = unknown
)

// variables that defines custom app templates.
var (
	helpHeaderTemplate = `{{template "helpNameTemplate" .}}

Usage: {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[options]{{end}}{{if .Commands}} command [command options]{{end}}{{end}}{{ if .Description}}

{{wrap .Description 0}}{{end}}`

	rootCommandTemplate = `%s

For listing options and commands, use '{{.HelpName}} --help or {{.HelpName}} -h'.
`

	appHelpTemplate = `%s{{if .VisibleFlags}}

Options: {{template "visibleFlagTemplate" .}}{{end}}{{if .VisibleCommands}}

Commands:{{template "visibleCommandCategoryTemplate" .}}{{end}}

For more information on a command, use '{{.HelpName}} [command] --help'.
`

	commandHelpTemplate = `{{.HelpName}}{{if .Description}} - {{template "descriptionTemplate" .}}{{end}}

Usage: {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}{{if .Args}}[arguments...]{{end}}[arguments...]{{end}}{{end}}{{if .VisibleFlags}}

Options: {{template "visibleFlagTemplate" .}}{{end}}{{if .VisibleCommands}}

Commands:{{template "visibleCommandCategoryTemplate" .}}{{end}}

`
)

// newCmd creates cli application defining custom help templates and default values.
func newCmd() *cli.App {
	app := &cli.App{}
	app.Name = "lb"
	app.Usage = "balance your AWS lambda layers cross regions"
	app.Description = "Layer Balancer or 'lb' is a tool for balancing the layer version of your Lambda\nacross AWS regions, so each region has the same Lambda layer version."
	app.Version = fmt.Sprintf("%s, build: %s, os: %s", releaseVersion, releaseCommit, releaseOS)
	app.CustomAppHelpTemplate = fmt.Sprintf(appHelpTemplate, helpHeaderTemplate)
	app.HideHelpCommand = true
	app.Suggest = true

	app.EnableBashCompletion = true
	app.BashComplete = func(ctx *cli.Context) {
		for _, cmd := range ctx.App.Commands {
			_, _ = fmt.Fprintln(ctx.App.Writer, cmd.Name)
		}
	}

	app.Action = func(cc *cli.Context) error {
		tpl := fmt.Sprintf(rootCommandTemplate, helpHeaderTemplate)
		cli.HelpPrinterCustom(cc.App.Writer, tpl, cc.App, nil)

		return nil
	}

	app.Commands = commands(bumpCmd, verifyCmd)

	return app
}

// commands sets custom help templates and default values.
func commands(cmds ...*cli.Command) []*cli.Command {
	for _, cmd := range cmds {
		cmd.Usage = cmd.Description
		cmd.HideHelpCommand = true
		cmd.CustomHelpTemplate = commandHelpTemplate
	}

	return cmds
}
