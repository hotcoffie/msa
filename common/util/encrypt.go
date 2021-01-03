package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/pkg/errors"
)

type Encrypt struct {
	PublicKey *rsa.PublicKey
}

func NewEncrypt(publicKey string) (*Encrypt, error) {
	key := ""
	for i := 0; i < len(publicKey); i += 64 {
		end := i + 64
		if end > len(publicKey)-1 {
			end = len(publicKey)
		}
		key += publicKey[i:end] + "\n"
	}
	publicKey = fmt.Sprintf(`-----BEGIN PUBLIC KEY-----
	%s-----END PUBLIC KEY-----`, key)
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return nil, errors.New("公钥编码失败")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("公钥生成失败")
	}
	return &Encrypt{pub}, nil
}

func (e *Encrypt) RsaEncrypt(word string) (string, error) {
	rs, err := rsa.EncryptPKCS1v15(rand.Reader, e.PublicKey, []byte(word))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(rs), nil
}
