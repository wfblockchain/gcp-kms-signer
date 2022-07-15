package main

import (
	"context"
	"log"
	"math/big"
	"net"
	"testing"
	"time"

	pb "github.com/wfblockchain/gcp-kms-signer-dlt/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const rpcURL = "https://rpc.flashbots.net/"

func TestEthService(t *testing.T) {
	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		t.Fatal(err)
	}

	srv := grpc.NewServer()

	server, err := newServer(rpcURL)
	if err != nil {
		t.Fatal(err)
	}
	pb.RegisterEthClientServiceServer(srv, server)
	reflection.Register(srv)

	// TODO: handle potential error here
	go srv.Serve(listener)

	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := pb.NewEthClientServiceClient(conn)
	gasPriceResp, err := client.SuggestGasPrice(ctx, &pb.Empty{})
	if err != nil {
		t.Fatal(err)
	}

	var gasPrice big.Int
	err = gasPrice.UnmarshalText(gasPriceResp.GetBigIntBytes())
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("suggested gas price: %v", gasPrice.String())
}
