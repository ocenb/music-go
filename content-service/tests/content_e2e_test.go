package tests_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTimeout = 5 * time.Minute

func TestHealthE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	env := setupTestEnv(ctx, t)

	resp, err := env.HTTPClient.Get(env.BaseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHistoryEmptyE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	env := setupTestEnv(ctx, t)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, env.BaseURL+"/api/content/history?take=10", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer test-token")

	resp, err := env.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestInternalDeleteUserE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	env := setupTestEnv(ctx, t)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, env.BaseURL+"/api/content/internal/users/999", nil)
	require.NoError(t, err)
	req.Header.Set("X-Internal-Secret", testInternalSecret)

	resp, err := env.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
