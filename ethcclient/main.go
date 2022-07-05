package main

import (
	"context"
	"log"
	"math/big"
	"time"

	pb "github.com/wfblockchain/gcp-kms-signer-dlt/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := pb.NewEthClientServiceClient(conn)
	gasPriceResp, err := client.SuggestGasPrice(ctx, &pb.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	var gasPrice big.Int
	err = gasPrice.UnmarshalText(gasPriceResp.GetBigIntBytes())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("suggested gas price: %v", gasPrice.String())
}
