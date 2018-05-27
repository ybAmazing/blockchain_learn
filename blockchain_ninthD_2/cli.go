package main

import "fmt"
import "flag"
import "os"

type CLI struct {
	bc      *BlockChain
	utxoset *UTXOSet
}

func (cli *CLI) send(from, to string, amount int) {
	tx := NewUTXOTransaction(from, to, amount, cli.bc, cli.utxoset)

	block := cli.bc.MineBlock([]*Transaction{tx})
	cli.utxoset.Update(block)
	cli.utxoset.PersistUTXOSet()

	fmt.Printf("Success send %d coins from %s to %s\n", amount, from, to)
}

func (cli *CLI) getBalance(address string) {
	balance := cli.utxoset.GetBalance(address)

	fmt.Printf("balance of address %s is %d coins.\n", address, balance)
}

func (cli *CLI) printChain() {
	bci := NewBlockchainIterator(cli.bc)

	for {
		block := bci.Next()

		PrintBlockInfo(block)

		if len(bci.currentHash) == 0 {
			break
		}
	}
}

func (cli *CLI) printUTXOSet() {
	utxoSet := cli.utxoset.UTXOSet

	for outPubKeyStr, utxos := range utxoSet {
		fmt.Printf("number of utxos of %s: %d\n", outPubKeyStr, len(utxos))

		for _, utxo := range utxos {
			fmt.Printf("    transaction str : %s\n", utxo.TxStr)
			fmt.Printf("    ouput index : %d\n", utxo.OutInd)
		}

		fmt.Println("----------------------------------")
	}

	// utxosetDB := UTXOSet{"NFC_UTXOset", "utxoset", make(map[string][]UTXO)}

	// utxosetDB.UTXOSet = cli.bc.GetUTXOSet()
	// for txstr, utxos := range utxosetDB.UTXOSet {
	// 	fmt.Printf("txstr of these utxos: %s\n", txstr)

	// 	for _, utxo := range utxos {
	// 		fmt.Printf("    utxo index : %d\n", utxo.OutInd)
	// 	}

	// 	fmt.Println("----------------------------------")
	// }

	// utxosetDB.PersistUTXOSet()

	// utxosetDB.LoadUTXOSet()
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage: [addblock -data='***'] [printchain]")
}

func (cli *CLI) Run() {
	// cli.validateArgs()
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendTxCmd := flag.NewFlagSet("send", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	printutxoset := flag.NewFlagSet("printutxoset", flag.ExitOnError)

	sendFrom := sendTxCmd.String("from", "", "the sender of this transaction")
	sendTo := sendTxCmd.String("to", "", "the recipetor of this transaction")
	sendAmount := sendTxCmd.Int("amount", 0, "amount of coin")

	getBlcAddr := getBalanceCmd.String("address", "", "which address do you want to query?")

	switch os.Args[1] {
	case "printchain":
		_ = printChainCmd.Parse(os.Args[2:])
	case "send":
		_ = sendTxCmd.Parse(os.Args[2:])
	case "getbalance":
		_ = getBalanceCmd.Parse(os.Args[2:])
	case "printutxoset":
		_ = printutxoset.Parse(os.Args[2:])
	default:
		cli.printUsage()
		os.Exit(1)
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

	if printutxoset.Parsed() {
		cli.printUTXOSet()
	}
}
