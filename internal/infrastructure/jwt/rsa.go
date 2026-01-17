package jwt

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"kiwi-user/config"
	"os"

	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/xerror"
)

var (
	ErrRSAKeyNotExists = errors.New("RSA key not exists")
)

type RSA struct {
	publicKeyPath  string
	privateKeyPath string

	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey

	logger logger.ILogger
}

func NewRSA(logger logger.ILogger, config *config.Config) *RSA {
	return &RSA{
		logger:         logger,
		publicKeyPath:  config.JWT.PublicKeyPath,
		privateKeyPath: config.JWT.PrivateKeyPath,
	}
}

// Init : Init the rsa setting, generate new public private files or load from existing files
func (r *RSA) Init() error {

	_, privateKeyPathErr := os.Stat(r.privateKeyPath)

	_, publicKeyPathErr := os.Stat(r.publicKeyPath)

	if !os.IsNotExist(privateKeyPathErr) && !os.IsNotExist(publicKeyPathErr) {
		r.logger.Infof(context.Background(), "load %s", r.privateKeyPath)
		r.logger.Infof(context.Background(), "load %s", r.publicKeyPath)
		// load private and public key from file
		publicKey, err := r.loadPublicKeyFromFile(r.publicKeyPath)
		if err != nil {
			return xerror.Wrap(err)
		}

		privateKey, err := r.loadPrivateKeyFromFile(r.privateKeyPath)
		if err != nil {
			return xerror.Wrap(err)
		}

		r.publicKey = publicKey
		r.privateKey = privateKey

		return nil
	}

	return xerror.Wrap(ErrRSAKeyNotExists)
}

func (r *RSA) loadPublicKeyFromFile(filePath string) (*rsa.PublicKey, error) {
	keybuffer, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keybuffer)
	if block == nil {
		return nil, errors.New("public key error")
	}

	pubkeyinterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publickey := pubkeyinterface.(*rsa.PublicKey)
	return publickey, nil
}

func (r *RSA) loadPrivateKeyFromFile(filePath string) (*rsa.PrivateKey, error) {
	keybuffer, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(keybuffer))
	if block == nil {
		return nil, errors.New("private key error")
	}

	privatekey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("parse private key error")
	}

	return privatekey, nil
}

// func (r *RSA) writePublicKeyToFile(publicKey *rsa.PublicKey, filePath string) error {
// 	keybytes, err := x509.MarshalPKIXPublicKey(publicKey)
// 	if err != nil {
// 		return err
// 	}
// 	block := &pem.Block{
// 		Type:  "PUBLIC KEY",
// 		Bytes: keybytes,
// 	}
// 	file, err := os.Create(filePath)
// 	if err != nil {
// 		return err
// 	}
// 	err = pem.Encode(file, block)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (r *RSA) writePrivatekeyToFile(privateKey *rsa.PrivateKey, filePath string) error {
// 	var keybytes = x509.MarshalPKCS1PrivateKey(privateKey)
// 	block := &pem.Block{
// 		Type:  "RSA PRIVATE KEY",
// 		Bytes: keybytes,
// 	}
// 	file, err := os.Create(filePath)
// 	if err != nil {
// 		return err
// 	}
// 	err = pem.Encode(file, block)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (r *RSA) generateKey() (*rsa.PrivateKey, *rsa.PublicKey, error) {
// 	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	publickey := &privatekey.PublicKey
// 	return privatekey, publickey, nil
// }

// SignWithPrivtaeKey : Sign the src data with private key
func (r *RSA) SignWithPrivateKey(src []byte, hash crypto.Hash) (signed []byte, e error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				e = errors.New(x)
			case error:
				e = x
			default:
				e = errors.New("unknown panic")
			}
		}
	}()
	h := hash.New()
	h.Write(src)
	hashed := h.Sum(nil)
	signed, err := rsa.SignPKCS1v15(rand.Reader, r.privateKey, hash, hashed)
	if err != nil {
		return []byte{}, err
	}

	return signed, nil
}

// VerifySignWithPublicKey : verify the signed data with public key
func (r *RSA) VerifySignWithPublicKey(src, signed []byte, hash crypto.Hash) (e error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				e = errors.New(x)
			case error:
				e = x
			default:
				e = errors.New("unknown panic")
			}
		}
	}()
	h := hash.New()
	h.Write(src)
	hashed := h.Sum(nil)
	err := rsa.VerifyPKCS1v15(r.publicKey, hash, hashed, signed)
	if err != nil {
		return xerror.Wrap(err)
	}
	return nil
}

// GetPublicKey : Get the public key
func (r *RSA) GetPublicKey() (string, error) {
	keybytes, err := x509.MarshalPKIXPublicKey(r.publicKey)
	if err != nil {
		return "", err
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: keybytes,
	}
	keybuffer := pem.EncodeToMemory(block)
	return string(keybuffer), nil
}
