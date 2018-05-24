package main

import "os"
import "fmt"
import "github.com/boltdb/bolt"

type NFC_Wallets struct {
	Wallets map[string]*Wallet
}

var Nfc_wallets NFC_Wallets

func LoadWallets() {
	walletsFile := "nfc_wallets"

	Nfc_wallets.Wallets = make(map[string]*Wallet)
	db, _ := bolt.Open(walletsFile, 0600, nil)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte([]byte("wallets")))

		if b == nil {
			fmt.Println("none wallet existing in db.\n")
			os.Exit(1)
		} else {
			b.ForEach(func(k, v []byte) error {
				wallet := DeSerializeWallet(v[:])
				Nfc_wallets.Wallets[string(k)] = wallet

				// show the address of wallet
				fmt.Println(string(k))

				return nil
			})
		}
		return nil
	})
}

func (wallets *NFC_Wallets) GetPubKeyFromAddr(address string) []byte {
	if wallet, ok := Nfc_wallets.Wallets[address]; ok {
		return wallet.PublicKey
	} else {
		fmt.Println("you can't do this transaction.")
		os.Exit(1)
	}
	return nil
}
