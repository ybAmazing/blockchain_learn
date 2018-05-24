package main

import "fmt"

import "github.com/boltdb/bolt"

func main() {
	wallet := NewWallet()

	fmt.Printf("private key: %x\n", wallet.PrivateKey)
	fmt.Printf("public key: %x\n", wallet.PublicKey)
	fmt.Printf("public key hash: %x\n", HashPubKey(wallet.PublicKey))
	fmt.Printf("address : %s\n", wallet.GetAddress())

	buffer := wallet.SerializeWallet()

	fmt.Println("--------------------------------")

	wallet2 := DeSerializeWallet(buffer)

	fmt.Printf("private key: %x\n", wallet2.PrivateKey)
	fmt.Printf("public key: %x\n", wallet2.PublicKey)
	fmt.Printf("public key hash: %x\n", HashPubKey(wallet2.PublicKey))
	fmt.Printf("address : %s\n", wallet2.GetAddress())

	fmt.Println("adding key pair.")

	walletsFile := "nfc_wallets"

	db, _ := bolt.Open(walletsFile, 0600, nil)

	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte([]byte("wallets")))

		if b == nil {
			b, _ = tx.CreateBucket([]byte("wallets"))

			b.Put([]byte(wallet.GetAddress()), wallet.SerializeWallet())
			fmt.Printf("add a wallet to db, address is %s\n", string(wallet.GetAddress()))
		} else {
			b.Put([]byte(wallet.GetAddress()), wallet.SerializeWallet())
			fmt.Printf("add a wallet to db, address is %s\n", string(wallet.GetAddress()))
		}
		return nil
	})
}
