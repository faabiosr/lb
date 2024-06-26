/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	"github.com/faabiosr/lb/internal"
)

var bumpCmd = &cli.Command{
	Name:        "bump",
	Description: "bump layer to latest version across regions",
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

		pterm.Printf(
			"Bumping layer across regions: %s\n",
			pterm.Green(strings.Join(regions, ", ")),
		)

		l := internal.LoadLayer(cfg, name)

		spin, err := spinner(cc.App.Writer, "getting latest version...").Start()
		if err != nil {
			return err
		}

		greatest, err := l.GreatestVersion(cc.Context, regions)
		if err != nil {
			_ = spin.Stop()
			return err
		}

		pterm.Printf(
			"Greatest version %d in region %s\n",
			greatest.Number,
			greatest.Region,
		)

		if greatest.Number == 0 {
			_ = spin.Stop()
			return errors.New("there are no published versions")
		}

		_ = spin.Stop()

		multi, err := pterm.DefaultMultiPrinter.Start()
		if err != nil {
			return err
		}

		g, ctx := errgroup.WithContext(cc.Context)

		for _, region := range regions {
			w := multi.NewWriter()
			g.Go(func() error {
				spin, err := spinner(w, fmt.Sprintf("%s: starting...", region)).Start()
				if err != nil {
					return err
				}

				latest, err := l.LatestVersion(ctx, region)
				if err != nil {
					return err
				}

				for i := latest.Number; i < greatest.Number; i++ {
					current, err := l.FetchVersion(ctx, i+1, greatest.Region)
					if err != nil {
						return err
					}

					spin.UpdateText(fmt.Sprintf("%s: downloading version %d", region, current.Number))

					buf := &bytes.Buffer{}
					if err := l.DownloadVersion(ctx, current, buf); err != nil {
						return err
					}

					current.Region = region
					current.Content.File = buf.Bytes()

					spin.UpdateText(fmt.Sprintf("%s: publishing version %d", region, current.Number))

					if err := l.PublishVersion(ctx, current); err != nil {
						return err
					}
				}

				_ = spin.Stop()
				pterm.Fprint(w, pterm.Sprintf("%s: bump complete", region))

				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}

		_, err = multi.Stop()
		return err
	},
}
