package proxyreqsign

import (
	fmtEcdsa "crypto/ecdsa"
	"fmt"
	"time"

	"github.com/LaunchPad-Network/NetPeek/constant"
	"github.com/LaunchPad-Network/NetPeek/internal/logger"
	"github.com/LaunchPad-Network/NetPeek/internal/misc/crypto/ecdsa"

	"github.com/spf13/viper"
)

var log = logger.New("ProxyReq Sign")
var privKey *fmtEcdsa.PrivateKey
var pubKey *fmtEcdsa.PublicKey

func init() {
	loadKeys()
}

func loadKeys() {
	prik, err := ecdsa.ImportPrivateKeyHex(viper.GetString("authentication.privatekey"))
	if err != nil {
		log.Error("Failed to load ECDSA private key: ", err)
	}

	pubk, err := ecdsa.ImportPublicKeyHex(viper.GetString("authentication.publickey"))
	if err != nil {
		log.Error("Failed to load ECDSA public key: ", err)

		newPriv, err := ecdsa.GenerateKey()
		if err != nil {
			log.Error("Failed to generate new ECDSA key pair: ", err)
		} else {
			newPub := &newPriv.PublicKey
			privHex := ecdsa.ExportPrivateKeyHex(newPriv)
			pubHex := ecdsa.ExportPublicKeyHex(newPub)
			log.Info("Generated new ECDSA key pair for signing proxy requests.")
			log.Infof("Private Key: %s", privHex)
			log.Infof("Public Key: %s", pubHex)
		}

		log.Fatal("Cannot continue without valid ECDSA public key")
	}

	privKey = prik
	pubKey = pubk
}

type SignedProxyRequest struct {
	Query     string
	Ts        int64
	Signature string
}

func (spr *SignedProxyRequest) Verify() bool {
	if spr.Ts < time.Now().Unix()-int64(constant.ProxyReqSignValidityDuration) {
		return false
	}
	toVerify := fmt.Sprintf("q=%s,ts=%d", spr.Query, spr.Ts)
	return ecdsa.VerifyText(pubKey, toVerify, spr.Signature)
}

func Sign(q string) (*SignedProxyRequest, error) {
	ts := time.Now().Unix()
	toSign := fmt.Sprintf("q=%s,ts=%d", q, ts)
	sig := ecdsa.SignText(privKey, toSign)
	return &SignedProxyRequest{
		Query:     q,
		Ts:        ts,
		Signature: sig,
	}, nil
}
