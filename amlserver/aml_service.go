package amlserver

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	pb "github.com/wfblockchain/gcp-kms-signer-dlt/proto"
	"golang.org/x/exp/slices"
)

type server struct {
	pb.UnimplementedAMLServiceServer
	blockList []string
}

func NewServer() (*server, error) {
	w := make([]string, 5)
	w[0] = "0x707E9E8D30e50dacD5C8866b658a4363c92FDdF2"
	w[1] = "0x3A4bdd260b4f2F033a722a79e7ee4BF0539de73D"
	w[2] = "0x91e7cE2cf99EAd1C15eACAeA848f3bAB0Ae415f9"
	w[3] = "0xE081abb7d9e327E89A13e65B3e2B6fcAF2eCEB97"
	w[4] = "0x20bB82F2Db6FF52b42c60cE79cDE4C7094Ce133F"
	return &server{blockList: w}, nil
}

func (s *server) Check(ctx context.Context, request *pb.AMLReq) (*pb.AMLResp, error) {
	addressBytes := request.GetAddressBytes()
	address := common.BytesToAddress(addressBytes)
	if slices.Contains(s.blockList, address.Hex()) {
		return &pb.AMLResp{Block: true}, nil
	} else {
		return &pb.AMLResp{Block: false}, nil
	}
}
