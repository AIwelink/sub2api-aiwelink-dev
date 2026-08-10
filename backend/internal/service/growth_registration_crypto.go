package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

const growthRegistrationCipherPrefix = "v1:"

var growthRegistrationCipherAdditionalData = []byte("sub2api:growth-registration-session:v1")

type GrowthRegistrationCipher struct {
	aead cipher.AEAD
}

func NewGrowthRegistrationCipher(keyHex string) (*GrowthRegistrationCipher, error) {
	if len(keyHex) != 64 {
		return nil, errors.New("growth registration encryption key must be 64 hexadecimal characters")
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != 32 {
		return nil, errors.New("growth registration encryption key must encode exactly 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create growth registration cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create growth registration AEAD: %w", err)
	}
	return &GrowthRegistrationCipher{aead: aead}, nil
}

func (c *GrowthRegistrationCipher) Encrypt(plaintext string) (string, error) {
	if c == nil || c.aead == nil {
		return "", errors.New("growth registration cipher is not initialized")
	}
	if len(plaintext) == 0 || len(plaintext) > GrowthRegistrationSessionMaxBytes {
		return "", errors.New("growth registration session must contain between 1 and 64 bytes")
	}

	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate growth registration nonce: %w", err)
	}
	payload := c.aead.Seal(nonce, nonce, []byte(plaintext), growthRegistrationCipherAdditionalData)
	return growthRegistrationCipherPrefix + base64.RawURLEncoding.EncodeToString(payload), nil
}

func (c *GrowthRegistrationCipher) Decrypt(ciphertext string) (string, error) {
	if c == nil || c.aead == nil {
		return "", errors.New("growth registration cipher is not initialized")
	}
	if !strings.HasPrefix(ciphertext, growthRegistrationCipherPrefix) {
		return "", errors.New("unsupported growth registration ciphertext")
	}
	maxPayloadBytes := c.aead.NonceSize() + GrowthRegistrationSessionMaxBytes + c.aead.Overhead()
	maxCiphertextBytes := len(growthRegistrationCipherPrefix) + base64.RawURLEncoding.EncodedLen(maxPayloadBytes)
	if len(ciphertext) > maxCiphertextBytes {
		return "", errors.New("growth registration ciphertext is too long")
	}

	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(ciphertext, growthRegistrationCipherPrefix))
	if err != nil {
		return "", errors.New("invalid growth registration ciphertext")
	}
	nonceSize := c.aead.NonceSize()
	if len(payload) < nonceSize+c.aead.Overhead() {
		return "", errors.New("invalid growth registration ciphertext")
	}
	plaintext, err := c.aead.Open(
		nil,
		payload[:nonceSize],
		payload[nonceSize:],
		growthRegistrationCipherAdditionalData,
	)
	if err != nil {
		return "", errors.New("invalid growth registration ciphertext")
	}
	if len(plaintext) == 0 || len(plaintext) > GrowthRegistrationSessionMaxBytes {
		return "", errors.New("invalid growth registration plaintext")
	}
	return string(plaintext), nil
}
