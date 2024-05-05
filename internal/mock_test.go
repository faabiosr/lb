/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package internal

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// Type alias to lambda options
type mockOpts = func(*lambda.Options)

type mockSvc struct {
	GetLayerVersionFn     func() (*lambda.GetLayerVersionOutput, error)
	ListLayerVersionsFn   func(...mockOpts) (*lambda.ListLayerVersionsOutput, error)
	PublishLayerVersionFn func() (*lambda.PublishLayerVersionOutput, error)
}

var _ svc = &mockSvc{}

func (m *mockSvc) GetLayerVersion(context.Context, *lambda.GetLayerVersionInput, ...mockOpts) (*lambda.GetLayerVersionOutput, error) {
	if m.GetLayerVersionFn != nil {
		return m.GetLayerVersionFn()
	}

	return &lambda.GetLayerVersionOutput{
		Content: &types.LayerVersionContentOutput{
			Location: aws.String(""),
		},
	}, nil
}

func (m *mockSvc) ListLayerVersions(_ context.Context, _ *lambda.ListLayerVersionsInput, opts ...mockOpts) (*lambda.ListLayerVersionsOutput, error) {
	if m.ListLayerVersionsFn != nil {
		return m.ListLayerVersionsFn(opts...)
	}

	return &lambda.ListLayerVersionsOutput{
		LayerVersions: []types.LayerVersionsListItem{
			{
				Version:                 10,
				CompatibleArchitectures: []types.Architecture{types.ArchitectureX8664},
				CompatibleRuntimes:      []types.Runtime{types.RuntimePython312},
			},
		},
	}, nil
}

func (m *mockSvc) PublishLayerVersion(context.Context, *lambda.PublishLayerVersionInput, ...mockOpts) (*lambda.PublishLayerVersionOutput, error) {
	if m.PublishLayerVersionFn != nil {
		return m.PublishLayerVersionFn()
	}

	return &lambda.PublishLayerVersionOutput{}, nil
}

type mockWriter struct{}

var _ io.Writer = &mockWriter{}

func (w *mockWriter) Write(p []byte) (int, error) {
	return 0, errors.New("failed")
}

type mockResponder struct {
	TripFn func() (*http.Response, error)
}

var _ http.RoundTripper = &mockResponder{}

func (m *mockResponder) RoundTrip(r *http.Request) (*http.Response, error) {
	if m.TripFn != nil {
		return m.TripFn()
	}

	return &http.Response{
		Body: io.NopCloser(strings.NewReader("content")),
	}, nil
}
