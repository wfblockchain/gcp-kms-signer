package main

import (
	"context"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	ds "github.com/wfblockchain/gcp-kms-signer-dlt/digestsigner"
	"github.com/wfblockchain/gcp-kms-signer-dlt/walletsigner"
)

const rpcURL = "https://cloudflare-eth.com"
const sevaAddress = "0x4549f47920997A486e9986d2e3e4540230534A03"

var (
	cred = &ds.KMSCred{
		ProjectID:  "certain-math-353822",
		Location:   "us-east4",
		KeyRing:    "wf_test",
		Key:        "anvil_test_secp256k1",
		KeyVersion: "1",
	}
)

func main() {
	ctx := context.Background()

	kmsSigner, err := ds.NewKMSSigner(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	signer := walletsigner.NewSigner(kmsSigner, 10*time.Second)
	log.Printf("signer accounts: %v", signer.Accounts())
	kmsPubkey, err := kmsSigner.GetPublicKey(ctx)
	if err != nil {
		log.Fatal(err)
	}

	kmsAddress := crypto.PubkeyToAddress(*kmsPubkey)
	log.Print(kmsPubkey, kmsAddress)

	//client, err := ethclient.Dial(rpcURL)
	//if err != nil {
	//	log.Fatal(err)
	//}
	// chainID, err := client.NetworkID(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//balance, err := client.BalanceAt(context.Background(), kmsAddress, nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("current balance: %v", balance)

	//nonce, err := client.PendingNonceAt(ctx, kmsAddress)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//value := big.NewInt(100)  // in wei NOTE: 1 eth = 10^{18} wei
	//gasLimit := uint64(21000) // in units
	//gasPrice, err := client.SuggestGasPrice(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//toAddress := common.HexToAddress(sevaAddress)
	//var data []byte
	//tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	//signedTx, err := signer.SignTx(signer.Accounts()[0], tx, chainID)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//err = client.SendTransaction(ctx, signedTx)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//log.Printf("tx sent: %s", tx.Hash())
}
