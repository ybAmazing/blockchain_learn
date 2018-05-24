package main

import "fmt"
import "strconv"
import "github.com/boltdb/bolt"
import "encoding/hex"

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func NewBlockchainIterator(bc *BlockChain) *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		block = DeSerializeBlock(b.Get(i.currentHash))

		return nil
	})

	i.currentHash = block.PreBlockHash

	return block
}

func PrintBlockInfo(block *Block) {

	fmt.Printf("previous hash : %x\n", block.PreBlockHash)
	fmt.Printf("block nonce ：%d\n", block.Nonce)
	fmt.Printf("block hash : %x\n", block.Hash)

	fmt.Printf("contains %d transactions\n", len(block.Transactions))
	for ind, tx := range block.Transactions {
		// fmt.Printf("	transaction id : %x\n", tx.ID)
		fmt.Printf("	transaction str : %s\n", hex.EncodeToString(tx.ID))
		fmt.Printf("	%d transaction contains %d input and %d output\n", ind, len(tx.Vin), len(tx.Vout))
		fmt.Printf("	is coinbase : %s\n", strconv.FormatBool(tx.IsCoinbase()))
		for outind, out := range tx.Vout {
			fmt.Printf("		the value of %d output : %d\n", outind, out.Value)
			fmt.Printf("		the pubkey hash of %d output : %x\n", outind, out.PubKeyHash)
		}

	}

	fmt.Printf("validate : %s\n\n", strconv.FormatBool(NewProofOfWork(block).Validate()))
}
