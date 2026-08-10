package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGrowthRegistrationSessionContext(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		value, ok := GrowthRegistrationSessionFromContext(context.Background())
		require.False(t, ok)
		require.Empty(t, value)
	})

	t.Run("present at byte limit", func(t *testing.T) {
		want := strings.Repeat("x", GrowthRegistrationSessionMaxBytes)
		ctx := WithGrowthRegistrationSession(context.Background(), want)
		value, ok := GrowthRegistrationSessionFromContext(ctx)
		require.True(t, ok)
		require.Equal(t, want, value)
	})

	t.Run("empty is ignored", func(t *testing.T) {
		ctx := WithGrowthRegistrationSession(context.Background(), "")
		_, ok := GrowthRegistrationSessionFromContext(ctx)
		require.False(t, ok)
	})

	t.Run("over limit is ignored", func(t *testing.T) {
		ctx := WithGrowthRegistrationSession(
			context.Background(),
			strings.Repeat("x", GrowthRegistrationSessionMaxBytes+1),
		)
		_, ok := GrowthRegistrationSessionFromContext(ctx)
		require.False(t, ok)
	})

	t.Run("nil context is safe", func(t *testing.T) {
		//nolint:staticcheck // Exercises the defensive nil-context fallback.
		require.Nil(t, WithGrowthRegistrationSession(nil, "growth-session"))
		//nolint:staticcheck // Exercises the defensive nil-context fallback.
		_, ok := GrowthRegistrationSessionFromContext(nil)
		require.False(t, ok)
	})
}
