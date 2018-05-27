package main

import "fmt"
import "os"
import "bytes"
import "github.com/boltdb/bolt"
import "encoding/hex"
import "encoding/gob"

type UTXOSet struct {
	dbFile     string
	bucketName string
	UTXOSet    map[string][]UTXO
}

// func (utxoset *UTXOSet) Init(dbFile, bucketName string) {
// 	gob.Register(UTXOSet{})
// 	utxoset.dbFile = dbFile
// 	utxoset.bucketName = bucketName
// }

func (bc *BlockChain) GetUTXOSet() map[string][]UTXO {
	bci := NewBlockchainIterator(bc)

	utxoSet := make(map[string][]UTXO)

	spentTxOutputs := make(map[string][]int)

	for {
		block := bci.Next()

		txs := block.Transactions

		for _, tx := range txs {
			txstr := hex.EncodeToString(tx.ID)

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					spentTxStr := hex.EncodeToString(in.Txid)
					spentTxOutputs[spentTxStr] = append(spentTxOutputs[spentTxStr], in.Vout)
				}
			}

		Outputs:
			for outInd, out := range tx.Vout {
				if spentTxOutputs[txstr] != nil {
					for _, spentOut := range spentTxOutputs[txstr] {
						if outInd == spentOut {
							continue Outputs
						}
					}
				}

				outPubKeyStr := hex.EncodeToString(out.PubKeyHash)
				utxoSet[outPubKeyStr] = append(utxoSet[outPubKeyStr], UTXO{txstr, outInd, out})

			}
		}

		if len(bci.currentHash) == 0 {
			break
		}
	}
	return utxoSet
}

func SerializeUTXOS(utxos []UTXO) []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	_ = encoder.Encode(utxos)

	return result.Bytes()
}

func DeserializeUTXOS(buffer []byte) []UTXO {
	utxos := []UTXO{}

	decoder := gob.NewDecoder(bytes.NewReader(buffer))

	err := decoder.Decode(&utxos)

	if err != nil {
		fmt.Println("Err is ", err)
		os.Exit(1)
	}

	return utxos
}

func (utxoset *UTXOSet) PersistUTXOSet() {
	db, _ := bolt.Open(utxoset.dbFile, 0600, nil)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoset.bucketName))

		if b == nil {
			b, _ = tx.CreateBucket([]byte(utxoset.bucketName))
		} else {
			_ = tx.DeleteBucket([]byte(utxoset.bucketName))
			b, _ = tx.CreateBucket([]byte(utxoset.bucketName))
		}

		for txstr, utxos := range utxoset.UTXOSet {
			txid, _ := hex.DecodeString(txstr)
			utxosBytes := SerializeUTXOS(utxos)
			err := b.Put(txid, utxosBytes)

			if err != nil {
				fmt.Println("Error is ", err)
				os.Exit(1)
			}
		}
		return nil
	})

	db.Close()
}

func (utxoset *UTXOSet) Update(block *Block) {
	for _, tx := range block.Transactions {
		for _, in := range tx.Vin {
			pubKeyHashStr := hex.EncodeToString(HashPubKey(in.PublicKey))
			txstr := hex.EncodeToString(in.Txid)

			newUTXOs := []UTXO{}
			for _, utxo := range utxoset.UTXOSet[pubKeyHashStr] {
				if utxo.TxStr == txstr && utxo.OutInd == in.Vout {
					continue
				} else {
					newUTXOs = append(newUTXOs, utxo)
				}
			}
			utxoset.UTXOSet[pubKeyHashStr] = newUTXOs
		}

		for outInd, out := range tx.Vout {
			pubKeyHashStr := hex.EncodeToString(out.PubKeyHash)

			utxoset.UTXOSet[pubKeyHashStr] = append(utxoset.UTXOSet[pubKeyHashStr], UTXO{hex.EncodeToString(tx.ID), outInd, out})
		}
	}

}

func LoadUTXOSet(dbFile, bucketName string) map[string][]UTXO {
	utxoset := make(map[string][]UTXO)

	db, _ := bolt.Open(dbFile, 0600, nil)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))

		if b == nil {
			fmt.Println("the UTXO set database doesn't exist.")
			os.Exit(1)
		}

		b.ForEach(func(k, v []byte) error {
			txstr := hex.EncodeToString(k)
			utxoset[txstr] = DeserializeUTXOS(v)
			return nil
		})
		return nil
	})

	db.Close()
	return utxoset
}
