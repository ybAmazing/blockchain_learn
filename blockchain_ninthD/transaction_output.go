package main

import "bytes"

type TxOutput struct {
	Value      int
	PubKeyHash []byte
	// ScriptPubKey string
}

func (out *TxOutput) CanBeUnlockedWith(address string) bool {
	// return out.ScriptPubKey == unlockingData

	pubKeyHash := GetPubKeyHashFromAddr(address)

	return bytes.Compare(pubKeyHash, out.PubKeyHash) == 0
}
