package main

import (
	ds "github.com/wfblockchain/gcp-kms-signer-dlt/digestsigner"
)

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

func main() {
}
