package main

import "fmt"
import "os"
import "time"
import "github.com/boltdb/bolt"
import "encoding/hex"

type BlockChain struct {
	//blocks []*Block
	tip []byte
	db  *bolt.DB
}

func NewBlockChain(rewardAddr string) *BlockChain {
	var tip []byte
	dbFile := "NFC_chain"

	db, _ := bolt.Open(dbFile, 0600, nil)

	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			coinbase := NewCoinbaseTx(rewardAddr, "")

			genesis := NewGenesis(coinbase)

			b, _ := tx.CreateBucket([]byte(blocksBucket))

			_ = b.Put(genesis.Hash, genesis.SerializeBlock())

			_ = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})

	bc := BlockChain{tip, db}

	return &bc
}

func (bc *BlockChain) AddBlock(transactions []*Transaction) *Block {
	newBlock := NewBlock(transactions, bc.tip)

	pow := NewProofOfWork(newBlock)

	nonce, hash := pow.Run()

	pow.block.Nonce = nonce
	pow.block.Hash = hash[:]

	_ = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		_ = b.Put(pow.block.Hash, pow.block.SerializeBlock())

		_ = b.Put([]byte("l"), pow.block.Hash)

		bc.tip = pow.block.Hash

		return nil
	})

	return newBlock
}

func NewGenesis(coinbase *Transaction) *Block {
	b := &Block{time.Now().Unix(), []*Transaction{coinbase}, []byte{}, []byte{}, 253, 0}

	pow := NewProofOfWork(b)
	nonce, hash := pow.Run()

	b.Nonce = nonce
	b.Hash = hash[:]

	return b
}

func (bc *BlockChain) MineBlock(transactions []*Transaction) *Block {
	for _, tx := range transactions {
		if tx.Verify(bc) == false {
			fmt.Println("found invalid transaction.")
			os.Exit(1)
		}
	}
	block := bc.AddBlock(transactions)
	fmt.Println("Success mint.")
	return block
}

func (utxoset *UTXOSet) FindUTXO(address string) []UTXO {
	pubKeyHashStr := hex.EncodeToString(GetPubKeyHashFromAddr(address))

	return utxoset.UTXOSet[pubKeyHashStr]
}

func (utxoset *UTXOSet) GetBalance(address string) int {
	utxos := utxoset.FindUTXO(address)
	balance := 0

	for _, utxo := range utxos {
		balance += utxo.Output.Value
	}
	return balance
}

func FindEnoughOutputs(from string, amount int, utxoset *UTXOSet) (int, []UTXO) {
	useUtxo := []UTXO{}
	sum := 0

	utxos := utxoset.FindUTXO(from)

	for _, out := range utxos {
		sum += out.Output.Value
		useUtxo = append(useUtxo, out)
		if sum >= amount {
			break
		}
	}

	return sum, useUtxo
}
