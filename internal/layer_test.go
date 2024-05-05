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
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

func TestFetchVersion(t *testing.T) {
	t.Run("response failure", func(t *testing.T) {
		l := &Layer{
			svc: &mockSvc{
				GetLayerVersionFn: func() (*lambda.GetLayerVersionOutput, error) {
					return nil, errors.New("failure")
				},
			},
		}

		_, err := l.FetchVersion(context.Background(), 1, "us-east-1")
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		l := &Layer{
			svc: &mockSvc{},
		}

		v, err := l.FetchVersion(context.Background(), 1, "us-east-1")
		if err != nil {
			t.Errorf("expected a nil, got error %v", err)
		}

		if v == nil {
			t.Error("expected a version, got nil")
		}
	})
}

func TestLatestVersion(t *testing.T) {
	t.Run("no version found", func(t *testing.T) {
		l := &Layer{
			svc: &mockSvc{
				ListLayerVersionsFn: func(...mockOpts) (*lambda.ListLayerVersionsOutput, error) {
					return &lambda.ListLayerVersionsOutput{}, nil
				},
			},
		}

		got, err := l.LatestVersion(context.Background(), "us-east-1")
		if err != nil {
			t.Errorf("expected nil, got error %v", err)
		}

		if got.Number != 0 {
			t.Errorf("expected version '0', got '%d'", got.Number)
		}
	})

	t.Run("latest version", func(t *testing.T) {
		l := &Layer{svc: &mockSvc{}}

		got, err := l.LatestVersion(context.Background(), "us-east-1")
		if err != nil {
			t.Errorf("expected nil, got error %v", err)
		}

		if got.Number != 10 {
			t.Errorf("expected version '10', got '%d'", got.Number)
		}

		if total := len(got.Architectures); total != 1 {
			t.Errorf("expected architectures '1', got '%d'", total)
		}

		if total := len(got.Runtimes); total != 1 {
			t.Errorf("expected runtimes '1', got '%d'", total)
		}
	})
}

func TestGreatestVersion(t *testing.T) {
	t.Run("response failure", func(t *testing.T) {
		l := &Layer{
			svc: &mockSvc{
				ListLayerVersionsFn: func(...mockOpts) (*lambda.ListLayerVersionsOutput, error) {
					return nil, errors.New("failure")
				},
			},
		}

		_, err := l.GreatestVersion(context.Background(), []string{"us-east-1"})
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("latest version", func(t *testing.T) {
		l := &Layer{svc: &mockSvc{}}

		got, err := l.GreatestVersion(context.Background(), []string{"us-east-1", "us-west-2"})
		if err != nil {
			t.Errorf("expected nil, got error %v", err)
		}

		if got.Number != 10 {
			t.Errorf("expected version '10', got '%d'", got.Number)
		}
	})
}

func TestDownloadVersion(t *testing.T) {
	tests := []struct {
		name   string
		svcFn  func() (*lambda.GetLayerVersionOutput, error)
		tripFn func() (*http.Response, error)
		writer io.Writer
		err    string
	}{
		{
			name:   "download request failure",
			writer: io.Discard,
			tripFn: func() (*http.Response, error) {
				return nil, errors.New("failed")
			},
			err: `failed to retrieve the layer content: Get "": failed`,
		},
		{
			name:   "download failure",
			writer: &mockWriter{},
			err:    "failed to download the layer version: failed",
		},
		{
			name:   "download success",
			writer: io.Discard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Layer{
				svc: &mockSvc{
					GetLayerVersionFn: tt.svcFn,
				},
				hc: &http.Client{Transport: &mockResponder{tt.tripFn}},
			}

			v := &Version{
				Number:  1,
				Region:  "us-east-1",
				Content: &Content{},
			}

			actual := l.DownloadVersion(context.Background(), v, tt.writer)

			var err string
			if actual != nil {
				err = actual.Error()
			}

			if tt.err != err {
				t.Errorf("Unexpected error: %s (expected %s)", err, tt.err)
			}

			if actual != nil {
				t.SkipNow()
			}
		})
	}
}

func TestPublishVersion(t *testing.T) {
	t.Run("nil version", func(t *testing.T) {
		l := &Layer{}

		err := l.PublishVersion(context.Background(), nil)
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("response failure", func(t *testing.T) {
		l := &Layer{
			svc: &mockSvc{
				PublishLayerVersionFn: func() (*lambda.PublishLayerVersionOutput, error) {
					return nil, errors.New("failure")
				},
			},
		}

		err := l.PublishVersion(context.Background(), &Version{Content: &Content{}})
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("publish success", func(t *testing.T) {
		l := &Layer{
			svc: &mockSvc{},
		}

		err := l.PublishVersion(context.Background(), &Version{Content: &Content{}})
		if err != nil {
			t.Errorf("expected nil, got error %v", err)
		}
	})
}
