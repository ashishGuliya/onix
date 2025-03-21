package signer

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/ashishGuliya/onix/pkg/log"
	"golang.org/x/crypto/blake2b"
)

// signer implements the signer interface and handles the signing process.
type signer struct {
}

// New creates a new Signer instance with the given configuration.
func New() (*signer, func() error, error) {
	s := &signer{}
	return s, nil, nil
}

// hash generates a signing string using BLAKE-512 hashing.
func hash(payload []byte, createdAt, expiresAt int64) (string, error) {
	hasher, _ := blake2b.New512(nil)

	_, err := hasher.Write(payload)
	if err != nil {
		return "", fmt.Errorf("failed to hash payload: %w", err)
	}

	hashSum := hasher.Sum(nil)
	digestB64 := base64.StdEncoding.EncodeToString(hashSum)

	return fmt.Sprintf("(created): %d\n(expires): %d\ndigest: BLAKE-512=%s", createdAt, expiresAt, digestB64), nil
}

// generateSignature signs the given signing string using the provided private key.
func generateSignature(signingString []byte, privateKeyBase64 string) ([]byte, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("error decoding private key: %w", err)
	}

	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key length")
	}

	privateKey := ed25519.PrivateKey(privateKeyBytes)
	return ed25519.Sign(privateKey, signingString), nil
}

// Sign generates a digital signature for the provided payload.
func (s *signer) Sign(ctx context.Context, body []byte, privateKeyBase64 string, createdAt, expiresAt int64) (string, error) {
	log.Debugf(ctx, "Attempting to sign with Key %s", privateKeyBase64)
	signingString, err := hash(body, createdAt, expiresAt)
	if err != nil {
		return "", err
	}

	signature, err := generateSignature([]byte(signingString), privateKeyBase64)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}
