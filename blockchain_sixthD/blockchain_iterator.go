package main

import "fmt"
import "strconv"
import "github.com/boltdb/bolt"

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
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

// this function is used to print info of block
func PrintBlockInfo(block *Block) {
	fmt.Printf("previous hash : %x\n", block.PreBlockHash)
	// fmt.Printf("block data : %s\n", block.Data)
	fmt.Printf("block nonce ：%d\n", block.Nonce)
	fmt.Printf("block hash : %x\n", block.Hash)
	fmt.Printf("validate : %s\n\n", strconv.FormatBool(NewProofOfWork(block).Validate()))
}
