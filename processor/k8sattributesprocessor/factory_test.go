// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package k8sattributesprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"
	"go.opentelemetry.io/collector/processor/xprocessor"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/metadata"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
}

func TestCreateProcessor(t *testing.T) {
	factory := NewFactory()

	realClient := kubeClientProvider
	kubeClientProvider = newFakeClient

	cfg := factory.CreateDefaultConfig()
	params := processortest.NewNopSettings(metadata.Type)

	tp, err := factory.CreateTraces(context.Background(), params, cfg, consumertest.NewNop())
	assert.NotNil(t, tp)
	assert.NoError(t, err)

	mp, err := factory.CreateMetrics(context.Background(), params, cfg, consumertest.NewNop())
	assert.NotNil(t, mp)
	assert.NoError(t, err)

	lp, err := factory.CreateLogs(context.Background(), params, cfg, consumertest.NewNop())
	assert.NotNil(t, lp)
	assert.NoError(t, err)

	pp, err := factory.(xprocessor.Factory).CreateProfiles(context.Background(), params, cfg, consumertest.NewNop())
	assert.NotNil(t, pp)
	assert.NoError(t, err)

	oCfg := cfg.(*Config)
	oCfg.Passthrough = true

	tp, err = factory.CreateTraces(context.Background(), params, cfg, consumertest.NewNop())
	assert.NotNil(t, tp)
	assert.NoError(t, err)

	mp, err = factory.CreateMetrics(context.Background(), params, cfg, consumertest.NewNop())
	assert.NotNil(t, mp)
	assert.NoError(t, err)

	lp, err = factory.CreateLogs(context.Background(), params, cfg, consumertest.NewNop())
	assert.NotNil(t, lp)
	assert.NoError(t, err)

	pp, err = factory.(xprocessor.Factory).CreateProfiles(context.Background(), params, cfg, consumertest.NewNop())
	assert.NotNil(t, pp)
	assert.NoError(t, err)

	// Switch it back so other tests run afterwards will not fail on unexpected state
	kubeClientProvider = realClient
}
