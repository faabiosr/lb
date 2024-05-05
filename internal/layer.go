/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package internal

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"golang.org/x/sync/errgroup"
)

// LoadLayer loads the layer information.
func LoadLayer(cfg aws.Config, name string) *Layer {
	return &Layer{
		Name: name,
		svc:  lambda.NewFromConfig(cfg),
		hc:   http.DefaultClient,
	}
}

type svc interface {
	GetLayerVersion(context.Context, *lambda.GetLayerVersionInput, ...func(*lambda.Options)) (*lambda.GetLayerVersionOutput, error)
	ListLayerVersions(context.Context, *lambda.ListLayerVersionsInput, ...func(*lambda.Options)) (*lambda.ListLayerVersionsOutput, error)
	PublishLayerVersion(context.Context, *lambda.PublishLayerVersionInput, ...func(*lambda.Options)) (*lambda.PublishLayerVersionOutput, error)
}

// Layer represents a lambda layer.
type Layer struct {
	Name string
	svc  svc
	hc   *http.Client
}

// Content represents the lambda layer content stored.
type Content struct {
	File     []byte
	Location string
}

// Version represents the lambda layer version.
type Version struct {
	Description   string
	Number        int64
	Region        string
	Content       *Content
	Architectures []types.Architecture
	Runtimes      []types.Runtime
	License       string
}

// withRegions is a helper option that sets Region on service requests.
func withRegion(r string) func(o *lambda.Options) {
	return func(o *lambda.Options) {
		o.Region = r
	}
}

// FetchVersion retrieves a version of lambda layer by region.
func (l *Layer) FetchVersion(ctx context.Context, version int64, region string) (*Version, error) {
	out, err := l.svc.GetLayerVersion(ctx, &lambda.GetLayerVersionInput{
		LayerName:     aws.String(l.Name),
		VersionNumber: aws.Int64(version),
	}, withRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve layer version: %w", err)
	}

	return &Version{
		Description: aws.ToString(out.Description),
		Number:      out.Version,
		Region:      region,
		Content: &Content{
			Location: aws.ToString(out.Content.Location),
		},
		Architectures: out.CompatibleArchitectures,
		Runtimes:      out.CompatibleRuntimes,
		License:       aws.ToString(out.LicenseInfo),
	}, nil
}

// LatestVersion retrieves the latest version of a lambda layer by region.
func (l *Layer) LatestVersion(ctx context.Context, region string) (*Version, error) {
	out, err := l.svc.ListLayerVersions(ctx, &lambda.ListLayerVersionsInput{
		LayerName: aws.String(l.Name),
		MaxItems:  aws.Int32(1),
	}, withRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to list layer versions: %w", err)
	}

	if len(out.LayerVersions) == 0 {
		return &Version{Region: region}, nil
	}

	latest := out.LayerVersions[0]

	return &Version{
		Description:   aws.ToString(latest.Description),
		Number:        latest.Version,
		Region:        region,
		Architectures: latest.CompatibleArchitectures,
		Runtimes:      latest.CompatibleRuntimes,
		License:       aws.ToString(latest.LicenseInfo),
	}, nil
}

// LatestVersions retrieves the latest version of all lambda layer regions.
func (l *Layer) LatestVersions(ctx context.Context, regions []string) ([]*Version, error) {
	versions := make([]*Version, len(regions))

	g, ctx := errgroup.WithContext(ctx)

	fn := func(index int, region string) func() error {
		return func() error {
			v, err := l.LatestVersion(ctx, region)
			if err != nil {
				return err
			}

			versions[index] = v

			return nil
		}
	}

	for i, r := range regions {
		g.Go(fn(i, r))
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("one of regions failed to retrieve the version: %w", err)
	}

	return versions, nil
}

// GreatestVersion retrieves the greatest version of the lambda layer across regions.
func (l *Layer) GreatestVersion(ctx context.Context, regions []string) (*Version, error) {
	versions, err := l.LatestVersions(ctx, regions)
	if err != nil {
		return nil, err
	}

	greatest := slices.MaxFunc(versions, func(c, n *Version) int {
		return cmp.Compare(c.Number, n.Number)
	})

	return greatest, nil
}

// DownloadVersion downloads the lambda layer version by region.
func (l *Layer) DownloadVersion(ctx context.Context, v *Version, w io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.Content.Location, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	res, err := l.hc.Do(req)
	if err != nil {
		return fmt.Errorf("failed to retrieve the layer content: %w", err)
	}
	defer res.Body.Close() // nolint:errcheck

	_, err = io.Copy(w, res.Body)
	if err != nil {
		return fmt.Errorf("failed to download the layer version: %w", err)
	}

	return nil
}

// PublishVersion publishes a new lambda layer version.
func (l *Layer) PublishVersion(ctx context.Context, v *Version) error {
	if v == nil {
		return errors.New("version must not be nil")
	}

	_, err := l.svc.PublishLayerVersion(ctx, &lambda.PublishLayerVersionInput{
		Content: &types.LayerVersionContentInput{
			ZipFile: v.Content.File,
		},
		LayerName:               aws.String(l.Name),
		CompatibleArchitectures: v.Architectures,
		CompatibleRuntimes:      v.Runtimes,
		Description:             aws.String(v.Description),
		LicenseInfo:             aws.String(v.License),
	}, withRegion(v.Region))
	if err != nil {
		return fmt.Errorf("failed to publish layer version: %w", err)
	}

	return err
}
