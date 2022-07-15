package main

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	pb "github.com/wfblockchain/gcp-kms-signer-dlt/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func TestAMLService(t *testing.T) {
	listener, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer()
	server, err := newServer()
	if err != nil {
		log.Fatal(err)
	}
	pb.RegisterAMLServiceServer(srv, server)
	reflection.Register(srv)

	// TODO: handle potential error here
	go srv.Serve(listener)

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
		t.Error(err)
	}
	if !resp.GetBlock() {
		t.Errorf("resp.GetBlock() = %v; want true", resp.GetBlock())
	}
	address = common.HexToAddress("0xddfabcdc4d8ffc6d5beaf154f18b778f892a0740")
	resp, err = client.Check(ctx, &pb.AMLReq{AddressBytes: address.Bytes()})
	if err != nil {
		log.Fatal(err)
	}
	if resp.GetBlock() {
		t.Errorf("resp.GetBlock() = %v; want false", resp.GetBlock())
	}
}
