package main

import "fmt"
import "bytes"
import "crypto/sha256"
import "time"
import "strconv"
import "math/big"
import "encoding/gob"
import "github.com/boltdb/bolt"
import "flag"
import "os"

const MaxNonce = 9999999

const blocksBucket = "blocks"

type Block struct {
	Timestamp    int64
	Data         []byte
	PreBlockHash []byte
	Hash         []byte
	Targetbits   uint
	Nonce        int
}

// struct and function relative to proof of work
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

type BlockChain struct {
	//blocks []*Block
	tip []byte
	db  *bolt.DB
}

func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)

	target.Lsh(target, block.Targetbits)

	pow := &ProofOfWork{block, target}

	return pow
}

func (pow *ProofOfWork) PrepareData(nonce int) []byte {
	data := bytes.Join([][]byte{pow.block.PreBlockHash, pow.block.Data, IntToHex(int64(pow.block.Timestamp)), IntToHex(int64(pow.block.Targetbits)), IntToHex(int64(nonce))}, []byte{})

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

// these functions used to persist the block to disk
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

func NewBlockChain() *BlockChain {
	var tip []byte
	dbFile := "NFC_chain"

	db, _ := bolt.Open(dbFile, 0600, nil)

	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := NewGenesis("this is a persisting genesis.")

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

// the fundamental functions relative to block chain
func NewBlock(data string, preBlockHash []byte) *Block {
	b := &Block{time.Now().Unix(), []byte(data), preBlockHash, []byte{}, 252, 0}

	pow := NewProofOfWork(b)
	nonce, hash := pow.Run()

	b.Nonce = nonce
	b.Hash = hash[:]

	return b
}

/*
// this function is used to add block without persisting blocks
func (bc *BlockChain) AddBlock(data string) {
	preBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, preBlock.Hash)

	pow := NewProofOfWork(newBlock)

	nonce, hash := pow.Run()

	pow.block.nonce = nonce
	pow.block.Hash = hash[:]

	bc.blocks = append(bc.blocks, pow.block)
}
*/

// this function is used to add block in persisting situation
func (bc *BlockChain) AddBlock(data string) {
	newBlock := NewBlock(data, bc.tip)

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

func NewGenesis(d string) *Block {
	b := &Block{time.Now().Unix(), []byte("this is genesis block"), []byte{}, []byte{}, 253, 0}

	pow := NewProofOfWork(b)
	nonce, hash := pow.Run()

	b.Nonce = nonce
	b.Hash = hash[:]

	return b
}

// this function used to iterate the block of the blockchain
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
	fmt.Printf("block data : %s\n", block.Data)
	fmt.Printf("block nonce ：%d\n", block.Nonce)
	fmt.Printf("block hash : %x\n", block.Hash)
	fmt.Printf("validate : %s\n\n", strconv.FormatBool(NewProofOfWork(block).Validate()))
}

// this struct type is used for CLI
type CLI struct {
	bc *BlockChain
}

//these functions are used as the command behavier of CLI
func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
}

func (cli *CLI) printChain() {
	bci := &BlockchainIterator{cli.bc.tip, cli.bc.db}

	for {
		block := bci.Next()

		PrintBlockInfo(block)

		if len(bci.currentHash) == 0 {
			break
		}
	}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage: [addblock -data='***'] [printchain]")
}

func (cli *CLI) Run() {
	// cli.validateArgs()
	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "addblock":
		_ = addBlockCmd.Parse(os.Args[2:])
	case "printchain":
		_ = printChainCmd.Parse(os.Args[2:])
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}

		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func main() {
	/*
		bc := &BlockChain{[]*Block{NewGenesis("this is genesis block")}}

		bc.AddBlock("first block")
		bc.AddBlock("second block")

		for _, block := range bc.blocks {
			fmt.Printf("previous hash : %x\n", block.PreBlockHash)
			fmt.Printf("block data : %s\n", block.Data)
			fmt.Printf("block nonce ：%d\n", block.nonce)
			fmt.Printf("block hash : %x\n", block.Hash)
			fmt.Printf("validate : %s\n\n", strconv.FormatBool(NewProofOfWork(block).Validate()))
		}
	*/

	bc := NewBlockChain()
	defer bc.db.Close()

	/*
		bc.AddBlock("this is first block")
		bc.AddBlock("this is second block")

		bci := &BlockchainIterator{bc.tip, bc.db}

		for {
			block := bci.Next()

			PrintBlockInfo(block)

			if len(bci.currentHash) == 0 {
				break
			}
		}
	*/

	cli := CLI{bc}

	cli.Run()
}
