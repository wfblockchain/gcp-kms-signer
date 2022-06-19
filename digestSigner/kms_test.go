package digestSigner

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

var (
	cred = &KMSCred{
		ProjectID:  "<account>",
		Location:   "<zone>",
		KeyRing:    "mock",
		Key:        "mock",
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
