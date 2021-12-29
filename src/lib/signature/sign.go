package lib_sig

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	eth "github.com/ethereum/go-ethereum/crypto"
	"github.com/iost-official/go-iost/account"
	crypto "github.com/iost-official/go-iost/crypto"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/nacl/sign"
	"golang.org/x/crypto/sha3"
)

func GenerateSignatureFromText(text string, privkey string) (string, error) {
	//if !IsBoardSolvable(board) {
	//	return "", fmt.Errorf("unsolvable board: %v", board)
	//}

	seckey, err := base58.Decode(privkey)
	if err != nil {
		return "", err
	}
	edKeypair, err := account.NewKeyPair(seckey, crypto.Ed25519)
	if err != nil {
		return "", err
	}
	digest := sha3.Sum256([]byte(text))
	fmt.Printf("info:%x\n", digest)

	buf := sign.Sign(nil, digest[:], (*[64]byte)(edKeypair.Seckey))
	signature := base58.Encode(buf[:64])

	return signature, nil
}

func IsLegalSig(fromEthAccount, sigHex string, msg []byte) bool {
	fromAddr := common.HexToAddress(fromEthAccount)

	sig := hexutil.MustDecode(sigHex)
	// https://github.com/ethereum/go-ethereum/blob/55599ee95d4151a2502465e0afc7c47bd1acba77/internal/ethapi/api.go#L442
	if sig[64] != 27 && sig[64] != 28 {
		return false
	}
	sig[64] -= 27

	pubKey, err := eth.SigToPub(signHash(msg), sig)
	if err != nil {
		return false
	}

	recoveredAddr := eth.PubkeyToAddress(*pubKey)

	return fromAddr == recoveredAddr
}

// https://github.com/ethereum/go-ethereum/blob/55599ee95d4151a2502465e0afc7c47bd1acba77/internal/ethapi/api.go#L404
// signHash is a helper function that calculates a hash for the given message that can be
// safely used to calculate a signature from.
//
// The hash is calculated as
//   keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
//
// This gives context to the signed message and prevents signing of transactions.
func signHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return eth.Keccak256([]byte(msg))
}
