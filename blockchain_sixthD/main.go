package main

import "strconv"

const MaxNonce = 9999999

const blocksBucket = "blocks"

func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}

func main() {
	bc := NewBlockChain("yanbiao")
	defer bc.db.Close()

	cli := CLI{bc}

	cli.Run()
}
