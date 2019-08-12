package bitmart

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ensure we're in sync with demo.py, from the bitmart documentation
func TestPrepareOrderSignature(t *testing.T) {
	auth := &Auth{
		key: APIKey{
			Secret: "api_secret", // obviously fake, but keeps us in track with demo output
		},
	}
	expect := "0ad8c5f31d66a1a7c70147146090a3c745ff330c766fd4e72352f6cfc736f860"
	got := prepareOrderSignature(auth, "BMX_ETH", "buy", 1.234, 1.5)
	require.Equal(t, expect, got)
}
