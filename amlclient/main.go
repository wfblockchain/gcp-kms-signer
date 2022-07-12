package main

import (
	"context"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	pb "github.com/wfblockchain/gcp-kms-signer-dlt/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	address := common.HexToAddress("0x707E9E8D30e50dacD5C8866b658a4363c92FDdF2")
	client := pb.NewAMLServiceClient(conn)
	resp, err := client.Check(ctx, &pb.AMLReq{AddressBytes: address.Bytes()})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%v is blocked: %v", address, resp.GetBlock())

	address = common.HexToAddress("0xddfabcdc4d8ffc6d5beaf154f18b778f892a0740")
	resp, err = client.Check(ctx, &pb.AMLReq{AddressBytes: address.Bytes()})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%v is blocked: %v", address, resp.GetBlock())
}
