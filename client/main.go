package main

import (
	"context"
	"log"
	"time"

	pb "github.com/wfblockchain/gcp-kms-signer-dlt/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewAddServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.Sign(ctx, &pb.SignTxReq{Tx: tx})
	if err != nil {
		log.Fatal(err)
	}
	log.Print(res.GetTx())
}
