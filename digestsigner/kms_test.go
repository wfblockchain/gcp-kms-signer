package digestsigner

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

var (
	cred = &KMSCred{
		ProjectID:  "certain-math-353822",
		Location:   "us-east4",
		KeyRing:    "wf_test",
		Key:        "anvil_test_secp256k1",
		KeyVersion: "1",
	}
)

func TestKMSSigner(t *testing.T) {
	ctx := context.Background()

	signer, err := NewKMSSigner(ctx, cred)
	if err != nil {
		t.Fatal(err)
	}
	for _, address := range signer.GetAddresses() {
		t.Logf("Signing digest using address %s", address)
		digest := crypto.Keccak256([]byte("test"))
		sig, err := signer.SignDigest(ctx, address, digest)
		if err != nil {
			t.Fatalf("failed to sign: %v", err)
		}
		t.Logf("got signature: %#x", sig)
		if !verifyDigest(address, digest, sig) {
			t.Fatalf("failed to verify signature")
		}
	}
}
