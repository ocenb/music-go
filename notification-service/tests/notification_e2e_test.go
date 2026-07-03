package tests_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTimeout = 5 * time.Minute

func TestEmailNotificationE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	env := setupTestEnv(ctx, t)

	t.Run("publish and consume", func(t *testing.T) {
		const (
			email = "user@example.com"
			msg   = "verification link"
		)

		publishNotification(ctx, t, env.Broker, env.Topic, email, msg)

		waitCtx, waitCancel := context.WithTimeout(ctx, 30*time.Second)
		defer waitCancel()

		recorded, err := env.smtp.waitForMessage(waitCtx)
		require.NoError(t, err)
		assert.Equal(t, email, recorded.To)
		assert.True(t, strings.Contains(recorded.Body, msg))
	})
}
