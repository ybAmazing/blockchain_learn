package main

import "bytes"

// import "fmt"

type TxInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PublicKey []byte
	// ScriptSig string
}

func (in *TxInput) CanUnlockOutputWith(address string) bool {

	// return bytes.Compare(Nfc_wallets.Wallets[address].PublicKey, in.PublicKey) == 0
	pubKeyHash := GetPubKeyHashFromAddr(address)
	return bytes.Compare(pubKeyHash, HashPubKey(in.PublicKey)) == 0
	// return in.ScriptSig == unlockingData
}
