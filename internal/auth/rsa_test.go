package auth_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/jaimesHub/bookmark-management/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeKeyPair generates RSA 2048 keypair, encodes PKCS#8 private + PKIX public
// to PEM, writes vào t.TempDir(). Returns 2 file paths.
// Match openssl genpkey + openssl rsa -pubout output format.
func writeKeyPair(t *testing.T) (privPath, pubPath string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "generate test RSA 2048 keypair")

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err, "marshal PKCS#8 private key")
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err, "marshal PKIX public key")
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	dir := t.TempDir()
	privPath = filepath.Join(dir, "private.pem")
	pubPath = filepath.Join(dir, "public.pem")
	require.NoError(t, os.WriteFile(privPath, privPEM, 0600))
	require.NoError(t, os.WriteFile(pubPath, pubPEM, 0644))
	return privPath, pubPath
}

// TC-L6-A001 — LoadKeys với valid RSA 2048 PEM pair → thành công.
func TestLoadKeys_ValidPair_Success(t *testing.T) {
	t.Parallel()
	privPath, pubPath := writeKeyPair(t)

	keys, err := auth.LoadKeys(privPath, pubPath)
	require.NoError(t, err)
	require.NotNil(t, keys)
	assert.NotNil(t, keys.Private)
	assert.NotNil(t, keys.Public)
	// Sanity: pub key match private's public component (round-trip integrity).
	assert.Equal(t, keys.Private.PublicKey.N.String(), keys.Public.N.String())
}

// TC-L6-A002 — Private key file không tồn tại → error wrapped với "load private key:".
func TestLoadKeys_PrivateMissing_Error(t *testing.T) {
	t.Parallel()
	_, pubPath := writeKeyPair(t)

	keys, err := auth.LoadKeys(filepath.Join(t.TempDir(), "nonexistent.pem"), pubPath)
	require.Error(t, err)
	assert.Nil(t, keys)
	assert.Contains(t, err.Error(), "load private key:", "error phải wrap với layer prefix")
}

// TC-L6-A003 — Public key file không tồn tại → error wrapped với "load public key:".
func TestLoadKeys_PublicMissing_Error(t *testing.T) {
	t.Parallel()
	privPath, _ := writeKeyPair(t)

	keys, err := auth.LoadKeys(privPath, filepath.Join(t.TempDir(), "nonexistent.pem"))
	require.Error(t, err)
	assert.Nil(t, keys)
	assert.Contains(t, err.Error(), "load public key:", "error phải wrap với layer prefix")
}

// TC-L6-A004 — File không phải PEM format → "invalid PEM" trong error chain.
func TestLoadKeys_InvalidPEMContent_Error(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	privPath := filepath.Join(dir, "private.pem")
	_, pubPath := writeKeyPair(t) // pub valid, để fail tới private parse

	require.NoError(t, os.WriteFile(privPath, []byte("not pem content"), 0600))

	keys, err := auth.LoadKeys(privPath, pubPath)
	require.Error(t, err)
	assert.Nil(t, keys)
	assert.Contains(t, err.Error(), "invalid PEM")
}

// TC-L6-A005 — EC private key (KHÔNG phải RSA) → "not an RSA private key" trong error chain.
func TestLoadKeys_NonRSAKey_Error(t *testing.T) {
	t.Parallel()

	// Generate ECDSA P-256 keypair → encode PKCS#8 PEM.
	ecPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	ecPrivBytes, err := x509.MarshalPKCS8PrivateKey(ecPriv)
	require.NoError(t, err)
	ecPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: ecPrivBytes})

	dir := t.TempDir()
	privPath := filepath.Join(dir, "ec-private.pem")
	require.NoError(t, os.WriteFile(privPath, ecPEM, 0600))

	// Pub key dùng valid RSA — chỉ private fail type check.
	_, pubPath := writeKeyPair(t)

	keys, err := auth.LoadKeys(privPath, pubPath)
	require.Error(t, err)
	assert.Nil(t, keys)
	assert.Contains(t, err.Error(), "not an RSA private key")
}
