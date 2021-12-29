package lib_iost

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCallAbi(t *testing.T) {
	// sdkインスタンスを作成.
	sdk, err := NewIOSTDevSDK()
	if err != nil {
		t.Fatal(err)
		return
	}
	// トークンを発行してみる.
	contractId := "some contract id"
	txHash, err := CallAbi(sdk, contractId, "issue", "", "admin", "[1]")
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Logf("txHash=%s", txHash)
	// トランザクションを取得.
	tx, err := sdk.GetTxByHash(txHash)
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Logf("Status=%v", tx.Status)
	t.Logf("TxReceipt.StatusCode=%v", tx.Transaction.TxReceipt.StatusCode)
	t.Logf("TxReceipt.Returns=%v", tx.Transaction.TxReceipt.Returns)
	t.Logf("TxReceipt.Receipts[0].Content=%v", tx.Transaction.TxReceipt.Receipts[0].Content)
	// トークンのメタデータを取得.
	funcName := fmt.Sprintf("%s/issue", contractId)
	for _, receipt := range tx.Transaction.TxReceipt.Receipts {
		if receipt.FuncName != funcName {
			continue
		}
		arr := []string{}
		json.Unmarshal([]byte(receipt.Content), &arr)
		tokenId := arr[1]
		metaData, err := GetTokenMetaData(sdk, tokenId, "some contract id")
		if err != nil {
			t.Fatal(err)
			return
		}
		t.Logf("metaData=%s", metaData)
	}
}

func TestMakeTransfer(t *testing.T) {
	iostSdkOpt := &IOSTDevSDKOptions{
		AccName: "mland_00107",
		SecKey:  "3rxeEk6y9LLNmECrHrTS3PmUscboogWKUkj3vQASZUbn7W8MTr4KPZoWUtj53Au9ik4W2apzbaMXgxuiZPW5Nkzk",
	}
	sdk, err := NewIOSTDevSDKWithOptions(iostSdkOpt)
	if err != nil {
		t.Error(err)
	}

	hash, err := CallAbi(sdk, "token.iost", "transfer", "jpya", iostSdkOpt.AccName, "mland_00445", "3", "i don't write memo")
	if err != nil {
		t.Error(err)
	}
	t.Log(hash)
}
