package main

import "fmt"
import "os"
import "bytes"
import "strconv"
import "encoding/hex"
import "crypto/ecdsa"
import "crypto/sha256"
import "crypto/rand"
import "math/big"
import "crypto/elliptic"

type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

type UTXO struct {
	TxStr  string
	OutInd int
	Output TxOutput
}

func NewCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txout := TxOutput{20, GetPubKeyHashFromAddr(to)}

	tx := &Transaction{[]byte{}, []TxInput{}, []TxOutput{txout}}

	tx.SetID()

	return tx
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 0
}

func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validUtxo := FindEnoughOutputs(from, amount, bc)

	if acc < amount {
		fmt.Println("balance isn't enough to pay for this transaction.")
		os.Exit(1)
	}

	for _, utxo := range validUtxo {
		txid, _ := hex.DecodeString(utxo.TxStr)

		txin := TxInput{txid, utxo.OutInd, []byte{}, Nfc_wallets.GetPubKeyFromAddr(from)}
		inputs = append(inputs, txin)
	}

	outputs = append(outputs, TxOutput{amount, GetPubKeyHashFromAddr(to)})
	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, GetPubKeyHashFromAddr(from)})
	}

	tx := &Transaction{[]byte{}, inputs, outputs}
	tx.SetID()
	tx.SetSignature(Nfc_wallets.Wallets[from].PrivateKey, bc)

	return tx
}

func (tx *Transaction) SetSignature(privatekey ecdsa.PrivateKey, bc *BlockChain) {
	if tx.IsCoinbase() {
		return
	}

	prevTx := tx.getPreviousTx(bc)

	for inInd, in := range tx.Vin {
		contentToSign := []byte{}

		// add the prev output info to the content to be signed
		contentToSign = bytes.Join([][]byte{in.Txid, []byte(strconv.Itoa(in.Vout))}, []byte{})

		// add the hash of the publickey of the sender to the content to be signed
		contentToSign = bytes.Join([][]byte{contentToSign, prevTx[hex.EncodeToString(in.Txid)].Vout[in.Vout].PubKeyHash}, []byte{})

		for _, out := range tx.Vout {
			// add the outputs info to the content to be signed
			contentToSign = bytes.Join([][]byte{contentToSign, out.PubKeyHash, []byte(strconv.Itoa(out.Value))}, []byte{})
		}

		hashToSign := sha256.Sum256(contentToSign)

		r, s, err := ecdsa.Sign(rand.Reader, &privatekey, hashToSign[:])
		if err != nil {
			fmt.Println("Error is ", err)
			os.Exit(-1)
		}

		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inInd].Signature = signature

	}
}

func (tx *Transaction) getPreviousTx(bc *BlockChain) map[string]Transaction {
	bci := NewBlockchainIterator(bc)

	prevTx := make(map[string]Transaction)

	for {
		block := bci.Next()

		for _, blockTx := range block.Transactions {
			for _, txIn := range tx.Vin {
				if bytes.Compare(txIn.Txid, blockTx.ID) == 0 {
					prevTx[hex.EncodeToString(txIn.Txid)] = *blockTx
				}
			}
		}

		if len(block.PreBlockHash) == 0 {
			break
		}
	}

	return prevTx
}

func (tx *Transaction) Verify(bc *BlockChain) bool {
	if tx.IsCoinbase() {
		return true
	}

	curve := elliptic.P256()
	prevTx := tx.getPreviousTx(bc)

	for _, in := range tx.Vin {
		contentToVerify := []byte{}
		contentToVerify = bytes.Join([][]byte{in.Txid, []byte(strconv.Itoa(in.Vout)), prevTx[hex.EncodeToString(in.Txid)].Vout[in.Vout].PubKeyHash}, []byte{})
		for _, out := range tx.Vout {
			contentToVerify = bytes.Join([][]byte{contentToVerify, out.PubKeyHash, []byte(strconv.Itoa(out.Value))}, []byte{})
		}

		hashToVerify := sha256.Sum256(contentToVerify)

		r := big.Int{}
		s := big.Int{}
		signLen := len(in.Signature)
		r.SetBytes(in.Signature[:signLen/2])
		s.SetBytes(in.Signature[signLen/2:])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PublicKey)
		x.SetBytes(in.PublicKey[:keyLen/2])
		y.SetBytes(in.PublicKey[keyLen/2:])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, hashToVerify[:], &r, &s) == false {
			return false
		}
	}
	return true
}
