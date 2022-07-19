package walletsignerserver

import (
	"context"
	"errors"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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
	amlServiceClient pb.AMLServiceClient
	ethServiceClient pb.EthServiceClient
	signer           *walletsigner.Signer
}

func NewServer() (*server, error) {
	ctx := context.Background()
	kmsSigner, err := digestsigner.NewKMSSigner(ctx, cred)
	if err != nil {
		return nil, err
	}
	signer := walletsigner.NewSigner(kmsSigner, 10*time.Second)
	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := pb.NewEthServiceClient(conn)

	conn2, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	aml_client := pb.NewAMLServiceClient(conn2)

	return &server{signer: &signer, ethServiceClient: client, amlServiceClient: aml_client}, nil
}

func (s *server) GetSignerAddress(ctx context.Context, request *pb.Empty) (*pb.SignerAddressResp, error) {
	address, err := s.signer.GetPubAddress(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.SignerAddressResp{AddressBytes: address.Bytes()}, nil
}

func (s *server) GetSignerPk(ctx context.Context, request *pb.Empty) (*pb.SignerPkResp, error) {
	pk, err := s.signer.GetPublicKey(ctx)
	if err != nil {
		return nil, err
	}
	pkBytes := crypto.FromECDSAPub(pk)
	return &pb.SignerPkResp{PkBytes: pkBytes}, nil
}

func (s *server) Sign(ctx context.Context, request *pb.SignTxReq) (*pb.SignTxResp, error) {
	marshalledTx := request.GetTx()
	tx := types.Transaction{}
	tx.UnmarshalBinary(marshalledTx)

	res, err := s.amlServiceClient.Check(ctx, &pb.AMLReq{AddressBytes: tx.To().Bytes()})
	if err != nil {
		return nil, err
	}
	if res.GetBlock() {
		return nil, errors.New("block listed address detected")
	}

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
