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
import "encoding/hex"
import "math/rand"

const MaxNonce = 9999999

const blocksBucket = "blocks"

type Block struct {
	Timestamp    int64
	Transactions []*Transaction
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

// the fundamental functions relative to block chain
func NewBlock(transactions []*Transaction, preBlockHash []byte) *Block {
	b := &Block{time.Now().Unix(), transactions, preBlockHash, []byte{}, 252, 0}

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
	// fmt.Printf("block data : %s\n", block.Data)
	fmt.Printf("block nonce ：%d\n", block.Nonce)
	fmt.Printf("block hash : %x\n", block.Hash)
	fmt.Printf("validate : %s\n\n", strconv.FormatBool(NewProofOfWork(block).Validate()))
}

// this struct type is used for CLI
type CLI struct {
	bc *BlockChain
}

//these functions are used as the command behavier of CLI
/*
func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
}
*/

func (cli *CLI) send(from, to string, amount int) {
	tx := NewUTXOTransaction(from, to, amount, cli.bc)

	cli.bc.MineBlock([]*Transaction{tx})

	fmt.Printf("Success send %d coins from %s to %s\n", amount, from, to)
}

func (cli *CLI) getBalance(address string) {
	balance := cli.bc.GetBalance(address)

	fmt.Printf("balance of address %s is %d coins.\n", address, balance)
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
	sendTxCmd := flag.NewFlagSet("send", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	sendFrom := sendTxCmd.String("from", "", "send TX from who")
	sendTo := sendTxCmd.String("to", "", "send TX to who")
	sendAmount := sendTxCmd.Int("amount", 0, "amount of coin")

	getBlcAddr := getBalanceCmd.String("address", "", "which address do you want to query?")

	switch os.Args[1] {
	case "addblock":
		_ = addBlockCmd.Parse(os.Args[2:])
	case "printchain":
		_ = printChainCmd.Parse(os.Args[2:])
	case "send":
		_ = sendTxCmd.Parse(os.Args[2:])
	case "getbalance":
		_ = getBalanceCmd.Parse(os.Args[2:])
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}

		// cli.AddBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendTxCmd.Parsed() {
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if getBalanceCmd.Parsed() {
		cli.getBalance(*getBlcAddr)
	}
}

// these struct types are used to completing transaction modual
type TxInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

type TxOutput struct {
	Value        int
	ScriptPubKey string
}

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

func (b *Block) HashTransaction() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func (tx *Transaction) SetID() []byte {
	hash := sha256.Sum256([]byte(strconv.Itoa(rand.Int())))
	return hash[:]
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

func (in *TxInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TxOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 0
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
							spentTxOutputs[txstr] = append(spentTxOutputs[txstr], in.Vout)
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
	bci := &BlockchainIterator{bc.tip, bc.db}

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
		if sum > amount {
			break
		}
	}

	return sum, useUtxo
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

func (bc *BlockChain) MineBlock(transactions []*Transaction) {
	bc.AddBlock(transactions)
	fmt.Println("Success mint.")
}

func main() {
	bc := NewBlockChain("yanbiao")
	defer bc.db.Close()

	cli := CLI{bc}

	cli.Run()
}
