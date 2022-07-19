package main

import (
	"context"
	"log"
	"math/big"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
const blockedAddress = "0x707E9E8D30e50dacD5C8866b658a4363c92FDdF2"

var (
	cred = &ds.KMSCred{
		ProjectID:  "certain-math-353822",
		Location:   "us-east4",
		KeyRing:    "wf_test",
		Key:        "anvil_test_secp256k1",
		KeyVersion: "1",
	}
)

func startETHService() {
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

func startWalletSignerService() {
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

func startAMLService() {
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

func startServices() {
	go startETHService()
	go startWalletSignerService()
	startAMLService()
}

func createETHServiceClient() pb.EthServiceClient {
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := pb.NewEthServiceClient(conn)
	return client
}

func createWalletSignerClient() pb.WalletServiceClient {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := pb.NewWalletServiceClient(conn)
	return client
}

func genDemoTx(ctx context.Context, ethServiceClient pb.EthServiceClient, toAddress string) (*types.Transaction, error) {
	gasPriceResp, err := ethServiceClient.SuggestGasPrice(ctx, &pb.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	var gasPrice big.Int
	err = gasPrice.UnmarshalText(gasPriceResp.GetBigIntBytes())
	if err != nil {
		return nil, err
	}

	walletSignerClient := createWalletSignerClient()
	resp, err := walletSignerClient.GetSignerAddress(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}

	nonceResp, err := ethServiceClient.PendingNonceAt(ctx, &pb.ECNonceReq{AddressBytes: resp.GetAddressBytes()})
	if err != nil {
		return nil, err
	}

	value := big.NewInt(100)  // in wei (1 eth = 10{^18} wei)
	gasLimit := uint64(21000) // in units

	var data []byte
	tx := types.NewTransaction(nonceResp.GetNonce(), common.HexToAddress(toAddress), value, gasLimit, &gasPrice, data)

	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	res, err := walletSignerClient.Sign(ctx, &pb.SignTxReq{Tx: txBytes})
	if err != nil {
		return nil, err
	}
	err = tx.UnmarshalBinary(res.GetTx())
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func runDemo() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ethServiceClient := createETHServiceClient()

	log.Printf("Create and then send TX to %v (allow listed)", common.HexToAddress(sevaAddress))
	tx, err := genDemoTx(ctx, ethServiceClient, sevaAddress)
	if err != nil {
		log.Fatal(err)
	}

	txBytes, err := tx.MarshalBinary()
	_, err = ethServiceClient.SendTx(ctx, &pb.ECTxReq{Tx: txBytes})
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(3 * time.Second)

	log.Printf("Create and then send TX to %v (block listed)", common.HexToAddress(blockedAddress))
	tx, err = genDemoTx(ctx, ethServiceClient, blockedAddress)
	if err != nil {
		log.Print(err)
		return
	}
}

func run() {
	go startServices()
	runDemo()
}

func main() {
	run()
}
