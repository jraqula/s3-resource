package s3resource

import (
	"encoding/pem"
	"crypto/x509"
	"crypto/rsa"
	"crypto"
	"errors"
	"fmt"
	"crypto/sha256"
	"crypto/rand"
)

type Signer interface {
	// Sign returns raw signature for the given data. This method
	// will apply the hash specified for the keytype to the data.
	Sign(data []byte) ([]byte, error)
}

type rsaPrivateKey struct {
	*rsa.PrivateKey
}

func ParsePrivateKey(pemBytes []byte) (Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("ssh: no key found")
	}

	var rawkey interface{}
	switch block.Type {
	case "RSA PRIVATE KEY":
		rsa, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rawkey = rsa
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %q", block.Type)
	}
	return newSignerFromKey(rawkey)
}

func newSignerFromKey(k interface{}) (Signer, error) {
	var sshKey Signer
	switch t := k.(type) {
	case *rsa.PrivateKey:
		sshKey = &rsaPrivateKey{t}
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %T", k)
	}
	return sshKey, nil
}

// Sign signs data with rsa-sha256
func (r *rsaPrivateKey) Sign(data []byte) ([]byte, error) {
	d := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, r.PrivateKey, crypto.SHA256, d[:])
}

// A Signer is can create signatures that verify against a public key.
type Verifier interface {
	// Sign returns raw signature for the given data. This method
	// will apply the hash specified for the keytype to the data.
	Verify(data[]byte, sig []byte) error
}

type rsaPublicKey struct {
	*rsa.PublicKey
}

// parsePublicKey parses a PEM encoded private key.
func ParsePublicKey(pemBytes []byte) (Verifier, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("ssh: no key found")
	}

	var rawkey interface{}
	switch block.Type {
	case "PUBLIC KEY":
		rsa, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rawkey = rsa
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %q", block.Type)
	}

	return newUnsignerFromKey(rawkey)
}

func newUnsignerFromKey(k interface{}) (Verifier, error) {
	var sshKey Verifier
	switch t := k.(type) {
	case *rsa.PublicKey:
		sshKey = &rsaPublicKey{t}
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %T", k)
	}
	return sshKey, nil
}

// Unsign verifies the message using a rsa-sha256 signature
func (r *rsaPublicKey) Verify(data []byte, signature []byte) error {
	d := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(r.PublicKey, crypto.SHA256, d[:], signature)
}