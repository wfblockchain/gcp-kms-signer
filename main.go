package main

import (
	"log"
	"net"

	as "github.com/wfblockchain/gcp-kms-signer-dlt/amlserver"
	ds "github.com/wfblockchain/gcp-kms-signer-dlt/digestsigner"
	es "github.com/wfblockchain/gcp-kms-signer-dlt/ethccserver"
	pb "github.com/wfblockchain/gcp-kms-signer-dlt/proto"
	ws "github.com/wfblockchain/gcp-kms-signer-dlt/walletsignerserver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const rpcURL = "https://rpc.flashbots.net/"
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

func start_eth_service() {
	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer()

	server, err := es.NewServer(rpcURL)
	if err != nil {
		log.Fatal(err)
	}
	pb.RegisterEthServiceServer(srv, server)
	reflection.Register(srv)

	if err := srv.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

func start_wallet_signer_service() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	server, err := ws.NewServer()
	if err != nil {
		log.Fatal(err)
	}
	pb.RegisterWalletServiceServer(srv, server)
	reflection.Register(srv)

	if err := srv.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

func start_aml_service() {
	listener, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer()
	server, err := as.NewServer()
	if err != nil {
		log.Fatal(err)
	}
	pb.RegisterAMLServiceServer(srv, server)
	reflection.Register(srv)

	if err := srv.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

func start_services() {
	go start_eth_service()
	go start_wallet_signer_service()
	start_aml_service()
}

// 1. Start ETH service
// 2. Start wallet signer service
// 3. Start AML service
func main() {
	start_services()
}
