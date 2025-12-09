package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/gin-gonic/gin"
)

func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func ExportPrivateKeyHex(key *ecdsa.PrivateKey) string {
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(der)
}

func ImportPrivateKeyHex(hexText string) (*ecdsa.PrivateKey, error) {
	der, err := hex.DecodeString(hexText)
	if err != nil {
		return nil, err
	}
	return x509.ParseECPrivateKey(der)
}

func ExportPublicKeyHex(pub *ecdsa.PublicKey) string {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(der)
}

func ImportPublicKeyHex(hexText string) (*ecdsa.PublicKey, error) {
	der, err := hex.DecodeString(hexText)
	if err != nil {
		return nil, err
	}
	pubKey, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, err
	}
	key, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not ECDSA public key")
	}
	return key, nil
}

func Sign(key *ecdsa.PrivateKey, msg []byte) string {
	r, s, _ := ecdsa.Sign(rand.Reader, key, msg)
	sig, _ := json.Marshal(gin.H{
		"R": r,
		"S": s,
	})
	return base64.StdEncoding.EncodeToString(sig)
}

func Verify(pub *ecdsa.PublicKey, msg []byte, sigText string) bool {
	sigBytes, err := base64.StdEncoding.DecodeString(sigText)
	if err != nil {
		return false
	}

	sig := struct {
		R big.Int
		S big.Int
	}{}
	err = json.Unmarshal(sigBytes, &sig)
	if err != nil {
		return false
	}

	var rInt, sInt = new(big.Int), new(big.Int)
	rInt.SetBytes(sig.R.Bytes())
	sInt.SetBytes(sig.S.Bytes())

	return ecdsa.Verify(pub, msg, rInt, sInt)
}

func SignText(key *ecdsa.PrivateKey, text string) string {
	return Sign(key, []byte(text))
}

func VerifyText(pub *ecdsa.PublicKey, text string, sigText string) bool {
	return Verify(pub, []byte(text), sigText)
}
