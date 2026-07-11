package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func ParseRSAPrivateKeyPEM(value []byte) (*rsa.PrivateKey, error) {
	block, rest := pem.Decode(value)
	if block == nil || len(rest) != 0 {
		return nil, ErrAccessTokenConfig
	}
	switch block.Type {
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, ErrAccessTokenConfig
		}
		privateKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, ErrAccessTokenConfig
		}
		return privateKey, nil
	case "RSA PRIVATE KEY":
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, ErrAccessTokenConfig
		}
		return privateKey, nil
	default:
		return nil, ErrAccessTokenConfig
	}
}

func ParseRSAPublicKeyPEM(value []byte) (*rsa.PublicKey, error) {
	block, rest := pem.Decode(value)
	if block == nil || len(rest) != 0 {
		return nil, ErrAccessTokenConfig
	}
	switch block.Type {
	case "PUBLIC KEY":
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, ErrAccessTokenConfig
		}
		publicKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, ErrAccessTokenConfig
		}
		return publicKey, nil
	case "RSA PUBLIC KEY":
		publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, ErrAccessTokenConfig
		}
		return publicKey, nil
	default:
		return nil, ErrAccessTokenConfig
	}
}
