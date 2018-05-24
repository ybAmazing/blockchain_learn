package main

import "strconv"

const MaxNonce = 9999999

const blocksBucket = "blocks"

func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}

func main() {
	LoadWallets()

	bc := NewBlockChain("LkGGzXxTNqvqVp34mgjrbz1qxuJ7yo9svg")
	defer bc.db.Close()

	utxoset := UTXOSet{"NFC_UTXOset", "utxoset", make(map[string][]UTXO)}

	utxoset.UTXOSet = bc.GetUTXOSet()
	utxoset.PersistUTXOSet()

	cli := CLI{bc}

	cli.Run()
}
