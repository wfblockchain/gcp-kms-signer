package walletsignerserver

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/wfblockchain/gcp-kms-signer-dlt/digestsigner"
	pb "github.com/wfblockchain/gcp-kms-signer-dlt/proto"
	"github.com/wfblockchain/gcp-kms-signer-dlt/walletsigner"
	"google.golang.org/grpc"
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
	ethServiceClient pb.EthServiceClient
	signer           *walletsigner.Signer
}

func newServer() (*server, error) {
	ctx := context.Background()
	kmsSigner, err := digestsigner.NewKMSSigner(ctx, cred)
	if err != nil {
		return nil, err
	}
	signer := walletsigner.NewSigner(kmsSigner, 10*time.Second)
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := pb.NewEthServiceClient(conn)
	resp, err := client.NetworkID(ctx, &pb.Empty{})
	if err != nil {
		log.Fatal(err)
	}
	log.Print(resp.GetBigIntBytes())
	return &server{signer: &signer, ethServiceClient: client}, nil
}

func (s *server) GetSignerAddress(ctx context.Context) common.Address {
	address, err := s.signer.GetPubAddress(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return address
}

func (s *server) GetSignerPk(ctx context.Context) (*ecdsa.PublicKey, error) {
	pk, err := s.signer.GetPublicKey(ctx)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

func (s *server) Sign(ctx context.Context, request *pb.SignTxReq) (*pb.SignTxResp, error) {
	marshalledTx := request.GetTx()
	tx := types.Transaction{}
	tx.UnmarshalBinary(marshalledTx)

	resp, err := s.ethServiceClient.NetworkID(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}
	var chainID big.Int
	chainID.UnmarshalText(resp.GetBigIntBytes())

	signedTx, err := s.signer.SignTx(s.signer.Accounts()[0], &tx, &chainID)
	if err != nil {
		return nil, err
	}
	marshalledTx, err = signedTx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return &pb.SignTxResp{Tx: marshalledTx}, nil
}
