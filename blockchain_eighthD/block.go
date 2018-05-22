package main

import "bytes"
import "crypto/sha256"
import "time"
import "strconv"
import "encoding/gob"
import "math/rand"

type Block struct {
	Timestamp    int64
	Transactions []*Transaction
	PreBlockHash []byte
	Hash         []byte
	Targetbits   uint
	Nonce        int
}

func (block *Block) SerializeBlock() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	_ = encoder.Encode(block)

	return result.Bytes()
}

func DeSerializeBlock(buffer []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(buffer))

	_ = decoder.Decode(&block)

	return &block
}

func NewBlock(transactions []*Transaction, preBlockHash []byte) *Block {
	b := &Block{time.Now().Unix(), transactions, preBlockHash, []byte{}, 252, 0}

	pow := NewProofOfWork(b)
	nonce, hash := pow.Run()

	b.Nonce = nonce
	b.Hash = hash[:]

	return b
}

func (b *Block) HashTransaction() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func (tx *Transaction) SetID() {
	rand.Seed(time.Now().UnixNano())
	hash := sha256.Sum256([]byte(strconv.Itoa(rand.Int())))
	tx.ID = hash[:]
}
