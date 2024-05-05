/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"errors"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/urfave/cli/v2"

	"github.com/faabiosr/lb/internal"
)

var verifyCmd = &cli.Command{
	Name:        "verify",
	Description: "verifies layer latest versions across regions",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:     "regions",
			Aliases:  []string{"r"},
			Usage:    "list of regions separated by comma.",
			Required: true,
		},
	},
	ArgsUsage: "layer-name",
	Action: func(cc *cli.Context) error {
		name := cc.Args().First()
		if name == "" {
			return errors.New(`required argument "layer-name" not set`)
		}

		regions := cc.StringSlice("regions")
		if len(regions) <= 1 {
			return errors.New(`required flag "regions" must contain at least two regions`)
		}

		cfg, err := config.LoadDefaultConfig(cc.Context)
		if err != nil {
			return fmt.Errorf("failed to load aws config: %w", err)
		}

		l := internal.LoadLayer(cfg, name)

		spin, err := spinner(cc.App.Writer, "verifying...").Start()
		if err != nil {
			return err
		}

		versions, err := l.LatestVersions(cc.Context, regions)
		if err != nil {
			_ = spin.Stop()
			return err
		}

		bumped := slices.CompactFunc(versions, func(c, v *internal.Version) bool {
			return c.Number == v.Number
		})

		if len(bumped) == 1 && bumped[0].Number == 0 {
			_ = spin.Stop()
			return errors.New("there are no published versions")
		}

		if len(bumped) > 1 {
			_ = spin.Stop()
			return errors.New(`some regions are not bumped`)
		}

		_ = spin.Stop()

		fmt.Fprintln(cc.App.Writer, "all regions bumped")
		return nil
	},
}
