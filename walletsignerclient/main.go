package main

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	pb "github.com/wfblockchain/gcp-kms-signer-dlt/proto"
	"google.golang.org/grpc"
)

const rpcURL = "https://cloudflare-eth.com"
const sevaAddress = "0x4549f47920997A486e9986d2e3e4540230534A03"

func genTestTx(ctx context.Context) (*types.Transaction, error) {
	kmsAddress := crypto.PubkeyToAddress(*kmsPK)

	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	nonce, err := ethClient.PendingNonceAt(ctx, kmsAddress)
	if err != nil {
		return nil, err
	}

	value := big.NewInt(100)  // in wei (1 eth = 10{^18} wei)
	gasLimit := uint64(21000) // in units
	gasPrice, err := ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	toAddress := common.HexToAddress(sevaAddress)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	return tx, nil
}

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewAddServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := genTestTx(ctx)
	if err != nil {
		log.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Sign(ctx, &pb.SignTxReq{Tx: txBytes})
	if err != nil {
		log.Fatal(err)
	}

	err = tx.UnmarshalBinary(res.GetTx())
	if err != nil {
		log.Fatal(err)
	}
	log.Print(tx.Hash())
}
