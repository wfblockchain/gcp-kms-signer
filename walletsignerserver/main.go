package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	ec "github.com/ethereum/go-ethereum/ethclient"
	"github.com/wfblockchain/gcp-kms-signer-dlt/digestsigner"
	pb "github.com/wfblockchain/gcp-kms-signer-dlt/proto"
	"github.com/wfblockchain/gcp-kms-signer-dlt/walletsigner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	cred = &digestsigner.KMSCred{
		ProjectID:  "certain-math-353822",
		Location:   "us-east4",
		KeyRing:    "wf_test",
		Key:        "anvil_test_secp256k1",
		KeyVersion: "1",
	}
)

type server struct {
	pb.UnimplementedWalletServiceServer
	ethclient *ec.Client
	signer    *walletsigner.Signer
}

func newServer() (*server, error) {

	ctx := context.Background()
	kmsSigner, err := digestsigner.NewKMSSigner(ctx, cred)
	if err != nil {
		return nil, err
	}
	signer := walletsigner.NewSigner(kmsSigner, 10*time.Second)
	return &server{signer: &signer, ethclient: client}, nil
}

func (s *server) Sign(ctx context.Context, request *pb.SignTxReq) (*pb.SignTxResp, error) {
	marshalledTx := request.GetTx()
	tx := types.Transaction{}
	tx.UnmarshalBinary(marshalledTx)

	chainID, err := s.ethclient.NetworkID(ctx)
	if err != nil {
		return nil, err
	}
	signedTx, err := s.signer.SignTx(s.signer.Accounts()[0], &tx, chainID)
	if err != nil {
		return nil, err
	}
	marshalledTx, err = signedTx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &pb.SignTxResp{Tx: marshalledTx}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer()

	server, err := newServer()
	if err != nil {
		log.Fatal(err)
	}
	pb.RegisterWalletServiceServer(srv, server)
	reflection.Register(srv)

	if e := srv.Serve(listener); e != nil {
		log.Fatal(err)
	}
}
