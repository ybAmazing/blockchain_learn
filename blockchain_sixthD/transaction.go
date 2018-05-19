package main

import "fmt"
import "os"
import "encoding/hex"

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

	txin := TxInput{[]byte{}, -1, data}
	txout := TxOutput{20, to}

	tx := &Transaction{[]byte{}, []TxInput{txin}, []TxOutput{txout}}
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

		txin := TxInput{txid, utxo.OutInd, from}
		inputs = append(inputs, txin)
	}

	outputs = append(outputs, TxOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	tx := &Transaction{[]byte{}, inputs, outputs}
	tx.SetID()

	return tx
}
