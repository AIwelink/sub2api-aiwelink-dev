package web

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbeddedFrontendBypassesReferralRoutes(t *testing.T) {
	require.True(t, shouldBypassEmbeddedFrontend("/r/abcdefgh"))
}
