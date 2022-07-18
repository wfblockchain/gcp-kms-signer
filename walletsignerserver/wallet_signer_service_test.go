package walletsignerserver

import (
	"context"
	"log"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	as "github.com/wfblockchain/gcp-kms-signer-dlt/amlserver"
	es "github.com/wfblockchain/gcp-kms-signer-dlt/ethccserver"
	pb "github.com/wfblockchain/gcp-kms-signer-dlt/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const rpcURL = "https://rpc.flashbots.net/"
const sevaAddress = "0x4549f47920997A486e9986d2e3e4540230534A03"

func serve() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	server, err := NewServer()
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

func genTestTx(ctx context.Context, address common.Address) (*types.Transaction, error) {
	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	nonce, err := ethClient.PendingNonceAt(ctx, address)
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

func TestWalletSignerService(t *testing.T) {
	go start_aml_service()
	go start_eth_service()
	go serve()

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := pb.NewWalletServiceClient(conn)

	resp, err := client.GetSignerAddress(ctx, &pb.Empty{})
	tx, err := genTestTx(ctx, common.BytesToAddress(resp.GetAddressBytes()))
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.Sign(ctx, &pb.SignTxReq{Tx: txBytes})
	if err != nil {
		t.Fatal(err)
	}
	err = tx.UnmarshalBinary(res.GetTx())
	if err != nil {
		t.Fatal(err)
	}
}
