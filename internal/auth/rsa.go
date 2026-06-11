// Package auth provides RSA keypair loading for JWT signing/verifying.
// LoadKeys eager-loads ở main bootstrap (fail-fast) — Lec-7 sẽ wire JWT layer.
package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// Keys aggregates loaded RSA keypair. Production main.go giữ instance này
// trong scope của engine để JWT sign (private) + verify (public).
type Keys struct {
	Private *rsa.PrivateKey
	Public  *rsa.PublicKey
}

// LoadKeys reads + parses RSA PEM files from disk.
// Returns wrapped error với layer prefix ("load private key:" hoặc "load public key:")
// để main.go log structured + caller distinguish failure mode.
//
// Plan § 4.8: openssl genpkey produces PKCS#8 → loadPrivate dùng ParsePKCS8PrivateKey.
// loadPublic dùng PKIX (SubjectPublicKeyInfo) — match openssl rsa -pubout output.
func LoadKeys(privPath, pubPath string) (*Keys, error) {
	priv, err := loadPrivate(privPath)
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}
	pub, err := loadPublic(pubPath)
	if err != nil {
		return nil, fmt.Errorf("load public key: %w", err)
	}
	return &Keys{Private: priv, Public: pub}, nil
}

// loadPrivate reads PKCS#8 PEM-encoded RSA private key.
// Returns "invalid PEM" nếu file content KHÔNG có PEM block.
// Returns "not an RSA private key" nếu key là EC/ECDSA/Ed25519.
func loadPrivate(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("invalid PEM")
	}
	// openssl genpkey -algorithm RSA → PKCS#8 wrapper; ParsePKCS1PrivateKey would fail.
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	return rsaKey, nil
}

// loadPublic reads PKIX (SubjectPublicKeyInfo) PEM-encoded RSA public key.
// openssl rsa -pubout produces PKIX format — match ParsePKIXPublicKey.
func loadPublic(path string) (*rsa.PublicKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("invalid PEM")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return rsaKey, nil
}
