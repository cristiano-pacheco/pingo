package registry

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/errs"
)

type PrivateKeyRegistry interface {
	Get() *rsa.PrivateKey
}

type registry struct {
	pk *rsa.PrivateKey
}

func NewPrivateKeyRegistry(conf config.Config) PrivateKeyRegistry {
	r := registry{}
	r.process(conf)
	return &r
}

func (r *registry) Get() *rsa.PrivateKey {
	return r.pk
}

func (r *registry) process(conf config.Config) {
	pkString, err := base64.StdEncoding.DecodeString(conf.JWT.PrivateKey)
	if err != nil {
		panic(err)
	}

	pk, err := mapPEMToRSAPrivateKey(pkString)
	if err != nil {
		panic(err)
	}

	r.pk = pk
}

func mapPEMToRSAPrivateKey(key []byte) (*rsa.PrivateKey, error) {
	var err error

	var block *pem.Block
	if block, _ = pem.Decode(key); block == nil {
		return nil, errs.ErrKeyMustBePEMEncoded
	}

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
			return nil, err
		}
	}

	var pkey *rsa.PrivateKey
	var ok bool
	if pkey, ok = parsedKey.(*rsa.PrivateKey); !ok {
		return nil, errs.ErrNotRSAPrivateKey
	}

	return pkey, nil
}
