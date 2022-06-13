## How does it work?

So we can trick the KMS to sign a message with a `Keccak256` digest.

**Please note there are some differences with the standard eth behavior:**

## How to use

 [`ec-sign-secp256k1-sha256`]

## Usage

Check out the code, it is simple.

`wallet_signer` implements the [`Wallet`](https://pkg.go.dev/github.com/ethereum/go-ethereum/accounts#Wallet) interface, and can be used as a wallet for eth libraries.

Or you can use `digest_singer` directly to sign a hashed data.
