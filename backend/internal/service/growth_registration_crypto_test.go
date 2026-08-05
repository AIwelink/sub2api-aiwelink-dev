package service

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const growthRegistrationTestKey = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

func TestGrowthRegistrationCipherRoundTrip(t *testing.T) {
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)

	first, err := cipher.Encrypt("growth-session-123")
	require.NoError(t, err)
	second, err := cipher.Encrypt("growth-session-123")
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(first, "v1:"))
	require.NotEqual(t, first, second, "each encryption must use a fresh nonce")
	require.NotContains(t, first, "growth-session-123")

	plaintext, err := cipher.Decrypt(first)
	require.NoError(t, err)
	require.Equal(t, "growth-session-123", plaintext)
}

func TestGrowthRegistrationCipherRequiresExactly32ByteHexKey(t *testing.T) {
	for name, key := range map[string]string{
		"empty":              "",
		"too short":          strings.Repeat("0", 63),
		"too long":           strings.Repeat("0", 66),
		"invalid hex":        strings.Repeat("z", 64),
		"thirty one bytes":   strings.Repeat("01", 31),
		"thirty three bytes": strings.Repeat("01", 33),
	} {
		t.Run(name, func(t *testing.T) {
			cipher, err := NewGrowthRegistrationCipher(key)
			require.Error(t, err)
			require.Nil(t, cipher)
		})
	}
}

func TestGrowthRegistrationCipherRejectsInvalidPlaintextLength(t *testing.T) {
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)

	_, err = cipher.Encrypt("")
	require.Error(t, err)
	_, err = cipher.Encrypt(strings.Repeat("x", GrowthRegistrationSessionMaxBytes+1))
	require.Error(t, err)

	encoded, err := cipher.Encrypt(strings.Repeat("x", GrowthRegistrationSessionMaxBytes))
	require.NoError(t, err)
	decoded, err := cipher.Decrypt(encoded)
	require.NoError(t, err)
	require.Len(t, decoded, GrowthRegistrationSessionMaxBytes)
}

func TestGrowthRegistrationCipherRejectsTampering(t *testing.T) {
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)
	ciphertext, err := cipher.Encrypt("growth-session-123")
	require.NoError(t, err)

	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(ciphertext, "v1:"))
	require.NoError(t, err)
	payload[len(payload)-1] ^= 0xff
	tampered := "v1:" + base64.RawURLEncoding.EncodeToString(payload)

	_, err = cipher.Decrypt(tampered)
	require.Error(t, err)
	_, err = cipher.Decrypt("v2:" + strings.TrimPrefix(ciphertext, "v1:"))
	require.Error(t, err)
}

func TestGrowthRegistrationCipherRejectsOversizedCiphertextBeforeDecode(t *testing.T) {
	cipher, err := NewGrowthRegistrationCipher(growthRegistrationTestKey)
	require.NoError(t, err)

	_, err = cipher.Decrypt("v1:" + strings.Repeat("A", 1024*1024))
	require.ErrorContains(t, err, "too long")
}

func TestGrowthRegistrationCipherRejectsUninitializedReceiver(t *testing.T) {
	var cipher *GrowthRegistrationCipher

	_, err := cipher.Encrypt("growth-session-123")
	require.Error(t, err)
	_, err = cipher.Decrypt("v1:invalid")
	require.Error(t, err)
}
