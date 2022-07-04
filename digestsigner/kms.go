package digestsigner

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type KMSCred struct {
	ProjectID   string
	Location    string
	KeyRing     string
	Key         string
	KeyVersion  string             // (Optional) if you want to use a specific key version
	TokenSource oauth2.TokenSource // (Optional) if you want to use a custom token source, e.g. a service account
}

func (c *KMSCred) keyname() string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", c.ProjectID, c.Location, c.KeyRing, c.Key)
}

func (c *KMSCred) keyversion() string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s/cryptoKeyVersions/%s", c.ProjectID, c.Location, c.KeyRing, c.Key, c.KeyVersion)
}

type KMSSigner struct {
	client           *kms.KeyManagementClient
	resourcePath     string
	addressVerionMap map[common.Address]string
	kmsCred          *KMSCred
}

func NewKMSSigner(ctx context.Context, cfg *KMSCred) (*KMSSigner, error) {
	client, err := kms.NewKeyManagementClient(ctx, option.WithTokenSource(cfg.TokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create kms client: %w", err)
	}
	s := &KMSSigner{
		client:           client,
		addressVerionMap: map[common.Address]string{},
		kmsCred:          cfg,
	}
	if err := s.loadAddress(ctx); err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}
	if len(s.addressVerionMap) == 0 {
		return nil, errors.New("no valid eth private key found")
	}
	return s, nil
}

func (s *KMSSigner) getPublicKey(ctx context.Context, key string) (*ecdsa.PublicKey, error) {
	resp, err := s.client.GetPublicKey(ctx, &kmspb.GetPublicKeyRequest{
		Name: key,
	})
	if err != nil {
		return nil, err
	}
	pk, err := pemToPubkey(resp.Pem)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

// GetPubAddress fetches the PEM and converts to canonical ETH address
func (s *KMSSigner) GetPubAddress(ctx context.Context) (common.Address, error) {
	pk, err := s.GetPublicKey(ctx)
	var address common.Address
	if err != nil {
		return address, err
	}
	return crypto.PubkeyToAddress(*pk), nil
}

// GetPublicKey fetches the PEM and converts to ecdsa public key format
func (s *KMSSigner) GetPublicKey(ctx context.Context) (*ecdsa.PublicKey, error) {
	if s.kmsCred.KeyVersion == "" {
		keyName := s.kmsCred.keyname()
		s.resourcePath = keyName
		it := s.client.ListCryptoKeyVersions(ctx, &kmspb.ListCryptoKeyVersionsRequest{
			Parent: keyName,
			Filter: "state=ENABLED AND algorithm=EC_SIGN_SECP256K1_SHA256",
		})
		for {
			resp, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, err
			}
			pk, err := s.getPublicKey(ctx, resp.GetName())
			if err != nil {
				return nil, err
			}
			return pk, nil

		}
	} else {
		s.resourcePath = s.kmsCred.keyversion()
		return s.getPublicKey(ctx, s.kmsCred.keyversion())
	}
	return nil, errors.New("no pubkey found")
}

func (s *KMSSigner) HasAddress(addr common.Address) bool {
	_, ok := s.addressVerionMap[addr]
	return ok
}

func (s *KMSSigner) ResourcePath() string {
	return s.resourcePath
}

func (s *KMSSigner) GetConnectionStatus() string {
	return s.client.Connection().GetState().String()
}

func (s *KMSSigner) GetAddresses() []common.Address {
	addresses := make([]common.Address, 0, len(s.addressVerionMap))
	for k := range s.addressVerionMap {
		addresses = append(addresses, k)
	}
	return addresses
}

func (s *KMSSigner) ListVersionedKeys() map[common.Address]string {
	result := map[common.Address]string{}
	for k, v := range s.addressVerionMap {
		result[k] = v
	}
	return result
}

func (s *KMSSigner) SignDigest(ctx context.Context, address common.Address, digest []byte) ([]byte, error) {
	keyVersion, ok := s.addressVerionMap[address]
	if !ok {
		return nil, fmt.Errorf("no eth private key found for address %s", address)
	}

	digestCRC32C := crc32c(digest)
	req := &kmspb.AsymmetricSignRequest{
		Name: keyVersion,
		Digest: &kmspb.Digest{
			Digest: &kmspb.Digest_Sha256{
				Sha256: digest,
			},
		},
		DigestCrc32C: wrapperspb.Int64(int64(digestCRC32C)),
	}

	// Call the API.
	result, err := s.client.AsymmetricSign(ctx, req)
	if err != nil {
		return nil, err
	}
	if !result.VerifiedDigestCrc32C {
		return nil, fmt.Errorf("AsymmetricSign: request corrupted in-transit")
	}

	if int64(crc32c(result.Signature)) != result.SignatureCrc32C.Value {
		return nil, fmt.Errorf("AsymmetricSign: response corrupted in-transit")
	}

	// recover R and S from the signature
	R, S, err := recoverRS(result.Signature)
	if err != nil {
		return nil, err
	}

	// Reconstruct the eth signature R || S || V
	sig := make([]byte, 65)
	copy(sig[:32], R.Bytes())
	copy(sig[32:64], S.Bytes())
	sig[64] = 0x1b

	// TODO: is ther a better way to determine the value of V?
	if !verifyDigest(address, digest, sig) {
		sig[64]++
		if !verifyDigest(address, digest, sig) {
			return nil, fmt.Errorf("AsymmetricSign: signature failed, unable to determine V")
		}
	}

	return sig, nil
}

func (s *KMSSigner) Close() error {
	return s.client.Close()
}

func (s *KMSSigner) loadAddress(ctx context.Context) error {
	if s.kmsCred.KeyVersion == "" {
		keyName := s.kmsCred.keyname()
		s.resourcePath = keyName
		it := s.client.ListCryptoKeyVersions(ctx, &kmspb.ListCryptoKeyVersionsRequest{
			Parent: keyName,
			Filter: "state=ENABLED AND algorithm=EC_SIGN_SECP256K1_SHA256",
		})
		for {
			resp, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			if err := s.setAddress(ctx, resp.GetName()); err != nil {
				return err
			}
		}
	} else {
		s.resourcePath = s.kmsCred.keyversion()
		return s.setAddress(ctx, s.kmsCred.keyversion())
	}
	return nil
}

func (s *KMSSigner) setAddress(ctx context.Context, key string) error {
	pk, err := s.getPublicKey(ctx, key)
	if err != nil {
		return err
	}
	s.addressVerionMap[crypto.PubkeyToAddress(*pk)] = key
	return nil
}
