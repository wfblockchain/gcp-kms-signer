package main

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	ds "github.com/wfblockchain/gcp-kms-signer-dlt/digestsigner"
	"github.com/wfblockchain/gcp-kms-signer-dlt/walletsigner"
)

const rpcURL = "https://cloudflare-eth.com"
const pem = `-----BEGIN PUBLIC KEY-----
MFYwEAYHKoZIzj0CAQYFK4EEAAoDQgAEK+pIyZ5c51/TQQVfikG86gzOdzpRP4vf
X0U93p2H9l6cw9acNdGoE9lVVPUp0/vMZ71ETrafJyWF7SwBcKg1GA==
-----END PUBLIC KEY-----`
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

	kmsPK, err := ds.PemToPubkey(pem)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("PEM to PK result: %v", kmsPK)

	kmsAddress := crypto.PubkeyToAddress(*kmsPK)
	log.Printf("PK to Address result: %v", kmsAddress)

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal(err)
	}

	nonce, err := client.PendingNonceAt(ctx, kmsAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(100)  // in wei (1 eth = 10{^18} wei)
	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress(sevaAddress)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		log.Fatal(err)
	}

	kmsSigner, err := ds.NewKMSSigner(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	signer := walletsigner.NewSigner(kmsSigner, 5)
	log.Printf("signer accounts: %v", signer.Accounts())
	signedTx, err := signer.SignTx(signer.Accounts()[0], tx, chainID)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("tx sent: %s", tx.Hash().Hex())
}
