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

func (bc *BlockChain) AddBlock(transactions []*Transaction) {
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
}

func NewGenesis(coinbase *Transaction) *Block {
	b := &Block{time.Now().Unix(), []*Transaction{coinbase}, []byte{}, []byte{}, 253, 0}

	pow := NewProofOfWork(b)
	nonce, hash := pow.Run()

	b.Nonce = nonce
	b.Hash = hash[:]

	return b
}

func (bc *BlockChain) MineBlock(transactions []*Transaction) {
	for _, tx := range transactions {
		if tx.Verify(bc) == false {
			fmt.Println("found invalid transaction.")
			os.Exit(1)
		}
	}
	bc.AddBlock(transactions)
	fmt.Println("Success mint.")
}

func (bci *BlockchainIterator) FindUTXO(address string) []UTXO {
	//unspentTxOutputs := make(map[string][]int)
	utxos := []UTXO{}
	spentTxOutputs := make(map[string][]int)

	for {
		block := bci.Next()

		txs := block.Transactions

		for _, tx := range txs {
			txstr := hex.EncodeToString(tx.ID)

		Outputs:
			for outInd, out := range tx.Vout {
				if spentTxOutputs[txstr] != nil {
					for _, spentOut := range spentTxOutputs[txstr] {
						if outInd == spentOut {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					// unspentTxOutputs[txid] = append(unspentTxOutputs[txid], outInd)
					utxos = append(utxos, UTXO{txstr, outInd, out})
				}

				if tx.IsCoinbase() == false {
					for _, in := range tx.Vin {
						if in.CanUnlockOutputWith(address) {
							spentTxStr := hex.EncodeToString(in.Txid)
							spentTxOutputs[spentTxStr] = append(spentTxOutputs[spentTxStr], in.Vout)
						}
					}
				}
			}
		}

		if len(bci.currentHash) == 0 {
			break
		}
	}
	return utxos
}

func (bc *BlockChain) GetBalance(address string) int {
	// bci := &BlockchainIterator{bc.tip, bc.db}
	bci := NewBlockchainIterator(bc)

	utxos := bci.FindUTXO(address)
	balance := 0

	for _, utxo := range utxos {
		balance += utxo.Output.Value
	}
	return balance
}

func FindEnoughOutputs(from string, amount int, bc *BlockChain) (int, []UTXO) {
	useUtxo := []UTXO{}
	sum := 0
	bci := BlockchainIterator{bc.tip, bc.db}

	utxos := bci.FindUTXO(from)

	for _, out := range utxos {
		sum += out.Output.Value
		useUtxo = append(useUtxo, out)
		if sum >= amount {
			break
		}
	}

	return sum, useUtxo
}
