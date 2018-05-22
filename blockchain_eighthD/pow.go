package main

import "bytes"
import "crypto/sha256"
import "math/big"

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)

	target.Lsh(target, block.Targetbits)

	pow := &ProofOfWork{block, target}

	return pow
}

func (pow *ProofOfWork) PrepareData(nonce int) []byte {
	data := bytes.Join([][]byte{pow.block.PreBlockHash, pow.block.HashTransaction(), IntToHex(int64(pow.block.Timestamp)), IntToHex(int64(pow.block.Targetbits)), IntToHex(int64(nonce))}, []byte{})

	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 1

	for nonce < MaxNonce {
		data := pow.PrepareData(nonce)

		hash := sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if pow.target.Cmp(&hashInt) == 1 {
			return nonce, hash[:]
		} else {
			nonce++
		}
	}

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.PrepareData(pow.block.Nonce)

	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
