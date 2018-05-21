package main

import "bytes"
import "crypto/sha256"
import "golang.org/x/crypto/ripemd160"
import "crypto/ecdsa"
import "crypto/elliptic"
import "encoding/gob"
import "crypto/rand"
import "github.com/btcsuite/btcutil/base58"
import "crypto/x509"
import "encoding/pem"

import "fmt"
import "os"

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

type SerializableWallet struct {
	PrivateKeyStr string
	PublicKey     []byte
}

func NewWallet() *Wallet {
	privateKey, publicKey := newKeyPair()
	wallet := &Wallet{privateKey, publicKey}

	return wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, _ := ecdsa.GenerateKey(curve, rand.Reader)
	public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, public
}

func (w Wallet) GetAddress() string {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte("1"), pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)

	address := base58.Encode(fullPayload)
	return address
}

func HashPubKey(pubkey []byte) []byte {
	pubKeySha256 := sha256.Sum256(pubkey)

	ripemd160Hasher := ripemd160.New()
	_, _ = ripemd160Hasher.Write(pubKeySha256[:])
	publicRIPEMD160 := ripemd160Hasher.Sum(nil)

	return publicRIPEMD160
}

func checksum(content []byte) []byte {
	firstSha256 := sha256.Sum256(content)
	secondSha256 := sha256.Sum256(firstSha256[:])

	return secondSha256[:4]
}

func GetPubKeyHashFromAddr(address string) []byte {
	decodeAddr := base58.Decode(address)

	pubKeyHash := decodeAddr[1 : len(decodeAddr)-4]

	return pubKeyHash[:]
}

func (wallet *Wallet) SerializeWallet() []byte {
	x509Encoded, _ := x509.MarshalECPrivateKey(&wallet.PrivateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	serializable := SerializableWallet{string(pemEncoded), wallet.PublicKey}
	result := serializable.SerializeHelper()

	return result[:]
}

func DeSerializeWallet(buffer []byte) *Wallet {
	sWallet := DeSerializeHelper(buffer)

	block, _ := pem.Decode([]byte(sWallet.PrivateKeyStr))
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	wallet := &Wallet{*privateKey, sWallet.PublicKey}

	return wallet
}

func (s_wallet *SerializableWallet) SerializeHelper() []byte {
	var result bytes.Buffer

	gob.Register(ecdsa.PrivateKey{})
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(s_wallet)

	if err != nil {
		fmt.Println("Error is ", err)
		os.Exit(-1)
	}

	return result.Bytes()
}

func DeSerializeHelper(buffer []byte) *SerializableWallet {
	var sWallet SerializableWallet

	decoder := gob.NewDecoder(bytes.NewReader(buffer))

	err := decoder.Decode(&sWallet)

	if err != nil {
		fmt.Println("Error is ", err)
		os.Exit(-1)
	}

	return &sWallet
}
