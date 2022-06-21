package main

import (
	"context"
	b64 "encoding/base64"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	ds "github.com/wfblockchain/gcp-kms-signer-dlt/digestsigner"
	ws "github.com/wfblockchain/gcp-kms-signer-dlt/walletsigner"
)

const rpcURL = "https://cloudflare-eth.com"
const b64EncodedPK = "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEeEKch3sXgwY8VEBXpORcMFRM4j6g1WFbjhvEX1jRv7xNNrDKSSBxk9m838HCJoLh8VBGVLyRPN1NVDi0HZXTfg=="
const sevaAddress = "0x4549f47920997A486e9986d2e3e4540230534A03"

// PubKeyAddr returns the Ethereum address for the (uncompressed) key bytes.
func pubKeyAddr(bytes []byte) common.Address {
	digest := crypto.Keccak256(bytes[1:])
	var addr common.Address
	copy(addr[:], digest[12:])
	return addr
}

var (
	cred = &ds.KMSCred{
		ProjectID:  "certain-math-353822",
		Location:   "us-east4",
		KeyRing:    "wf_test",
		Key:        "anvil_test",
		KeyVersion: "1",
	}
)

func main() {
	ctx := context.Background()

	decodedPK, _ := b64.StdEncoding.DecodeString(b64EncodedPK)
	pk := pubKeyAddr([]byte(decodedPK))
	log.Println(pk.String())

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal(err)
	}

	fromAddress := pk
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
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

	// TODO: use KMS signer here
	kmsSigner, err := ds.NewKMSSigner(context.Background(), cred)
	signer := ws.NewSigner(kmsSigner, 0)
	signedTx, err := signer.SignTx(pk, tx, chainID)

	// signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// TODO: get this to work with GCP KMS... getting a permissions error now
	// ctx := context.Background()
	// signer, err := ds.NewKMSSigner(ctx, cred)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for _, address := range signer.GetAddresses() {
	// 	log.Printf("Signing digest using address %s", address)
	// 	digest := crypto.Keccak256([]byte("test"))
	// 	sig, err := signer.SignDigest(ctx, address, digest)
	// 	if err != nil {
	// 		log.Fatalf("failed to sign: %v", err)
	// 	}
	// 	log.Printf("got signature: %#x", sig)
	// }

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("tx sent: %s", tx.Hash().Hex())
}
