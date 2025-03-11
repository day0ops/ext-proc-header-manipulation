package envoy_test

import (
	"context"
	"testing"

	"github.com/day0ops/ext-proc-header-manipulation/test/containers/envoy"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func TestRunContainer(t *testing.T) {
	container := envoy.NewTestContainer()
	err := container.Run(context.Background(), "quay.io/solo-io/envoy-gloo:1.34.0-patch0")
	defer testcontainers.CleanupContainer(t, container)

	require.NoError(t, err)
	require.Contains(t, container.URL.String(), "http://localhost:")
}
