package lib_iost

import (
	"fmt"
	"math/rand"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/sdk"
)

type createIostAccountResult struct {
	AccountName string
	PubKey      string
	SecKey      string
	TxHash      string
}

var accIdValidChars = []byte("abcdefghijklmnopqrstuvwxyz1234567890")

func CreateIostAccount(sdkIns *sdk.IOSTDevSDK, accName string) (*createIostAccountResult, error) {
	var err error
	keyPair, err := account.NewKeyPair(nil, crypto.NewAlgorithm("ed25519"))
	if err != nil {
		return nil, err
	}

	// Keys as string
	pubKey := common.Base58Encode(keyPair.Pubkey)
	secKey := common.Base58Encode(keyPair.Seckey)

	// Might as well check len here
	if len(accName) < 5 || len(accName) > 11 {
		return nil, fmt.Errorf("Invalid acc name length")
	}
	accRes := &createIostAccountResult{
		AccountName: accName,
		PubKey:      pubKey,
		SecKey:      secKey,
	}
	txHash, err := sdkIns.CreateNewAccount(accName, pubKey, pubKey, 10, 0, 0)
	accRes.TxHash = txHash
	return accRes, nil
}

func CreateIdolverseIostAccount(sdkIns *sdk.IOSTDevSDK) (*createIostAccountResult, error) {
	buf := make([]byte, 8)
	for i := range buf {
		buf[i] = accIdValidChars[rand.Intn(len(accIdValidChars))]
	}

	accName := fmt.Sprintf("pl_%s", string(buf))
	return CreateIostAccount(sdkIns, accName)
}

// Checks if an Account is the owner of a pubkey
func IsPubKeyOwner(sdkIns *sdk.IOSTDevSDK, accName string, pubKey string) (bool, error) {
	acc, err := sdkIns.GetAccountInfo(accName)
	if err != nil {
		return false, err
	}
	accPermissions := acc.GetPermissions()

	ownerPermissions, ok := accPermissions["owner"]
	if !ok {
		return false, nil
	}
	ownerAccItems := ownerPermissions.GetItems()

	pubKeyFound := false
	for _, accItem := range ownerAccItems {
		if accItem.IsKeyPair && accItem.Id == pubKey {
			pubKeyFound = true
		}
	}
	return pubKeyFound, nil
}
