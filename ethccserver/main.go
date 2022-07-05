package main

import (
	"context"
	"log"
	"net"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ec "github.com/ethereum/go-ethereum/ethclient"
	pb "github.com/wfblockchain/gcp-kms-signer-dlt/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const rpcURL = "https://rpc.flashbots.net/"

type server struct {
	pb.UnimplementedEthClientServiceServer
	ethclient *ec.Client
}

func newServer(rpcURL string) (*server, error) {
	client, err := ec.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return &server{ethclient: client}, nil
}

func (s *server) PendingNonceAt(ctx context.Context, request *pb.ECNonceReq) (*pb.ECNonceResp, error) {
	addressBytes := request.GetAddressBytes()
	kmsAddress := common.BytesToAddress(addressBytes)
	nonce, err := s.ethclient.PendingNonceAt(ctx, kmsAddress)
	if err != nil {
		return nil, err
	}
	return &pb.ECNonceResp{Nonce: nonce}, nil
}

func (s *server) SuggestGasPrice(ctx context.Context, request *pb.Empty) (*pb.ECGasPriceResp, error) {
	gasPrice, err := s.ethclient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("gas price on server side: %v", gasPrice)
	gasPriceBytes, err := gasPrice.MarshalText()
	if err != nil {
		return nil, err
	}
	return &pb.ECGasPriceResp{BigIntBytes: gasPriceBytes}, nil
}

func (s *server) SendTx(ctx context.Context, request *pb.ECTxReq) (*pb.Empty, error) {
	var signedTx types.Transaction
	err := signedTx.UnmarshalBinary(request.GetTx())
	if err != nil {
		return nil, err
	}
	err = s.ethclient.SendTransaction(ctx, &signedTx)
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer()

	server, err := newServer(rpcURL)
	if err != nil {
		log.Fatal(err)
	}
	pb.RegisterEthClientServiceServer(srv, server)
	reflection.Register(srv)

	if e := srv.Serve(listener); e != nil {
		log.Fatal(err)
	}
}
