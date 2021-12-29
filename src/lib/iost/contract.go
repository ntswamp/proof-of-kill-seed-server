package lib_iost

import (
	lib_error "app/src/lib/error"
	"encoding/json"
	"os"
	"strconv"

	rpcpb "github.com/iost-official/go-iost/rpc/pb"
	"github.com/iost-official/go-iost/sdk"
)

// =====================================
// コントラクトのストレージのキー.
// トークンの持ち主情報.
const TOKEN_OWNER = "tokenownerby"

// ユーザーの所持トークン情報.
const HOLD_TOKEN = "holdtoken"

// メタデータ.
const METADATA_TOKEN = "metadata#cl_item_token"

// =====================================

// コントラクトアカウント名.
var contractAccountName string = os.Getenv("CONTRACT_ACCOUNT_MAIN")

// コントラクトアカウントの秘密鍵.
var contractAccountKey string = os.Getenv("CONTRACT_ACCOUNT_KEY_MAIN")

func GetContractAccountName() string {
	return contractAccountName
}
func GetContractAccountKey() string {
	return contractAccountKey
}

/**
 * クロスリンクのコントラクト呼び出し.
 * 別途トランザクションを追跡して確認する必要がある.
 */
func CallAbi(sdkIns *sdk.IOSTDevSDK, contractId, abi string, args ...interface{}) (string, error) {
	// 引数をjsonに.
	b, err := json.Marshal(args)
	if err != nil {
		return "", lib_error.WrapError(err)
	}
	// コントラクト呼び出しのトランザクションを作成.
	action := sdk.NewAction(contractId, abi, string(b))
	txHash, err := sdkIns.SendTxFromActions([]*rpcpb.Action{action})
	if err != nil {
		return "", lib_error.WrapError(err)
	}
	return txHash, nil
}

/**
 * コントラクトのストレージデータを取得.
 */
func GetContractStorageData(sdkIns *sdk.IOSTDevSDK, contractId, dataKey, dataField string) (string, error) {
	req := &rpcpb.GetContractStorageRequest{
		Id:             contractId,
		Key:            dataKey,
		Field:          dataField,
		ByLongestChain: true,
	}
	response, err := sdkIns.GetContractStorage(req)
	if err != nil {
		return "", lib_error.WrapError(err)
	}
	return response.Data, nil
}

/**
 * トークンの持ち主アカウント名を取得する.
 */
func GetTokenOwner(sdkIns *sdk.IOSTDevSDK, tokenId string, contractId string) (string, error) {
	return GetContractStorageData(sdkIns, contractId, TOKEN_OWNER, tokenId)
}

/**
 * トークンのメタデータを取得する.
 */
func GetTokenMetaData(sdkIns *sdk.IOSTDevSDK, tokenId string, contractId string) (string, error) {
	return GetContractStorageData(sdkIns, contractId, METADATA_TOKEN, tokenId)
}

/**
 * トークンのメタデータを読み取って種類とIDを返す取得する.
 */
func GetTokenItemId(sdkIns *sdk.IOSTDevSDK, tokenId string, contractId string) (int, uint64, error) {
	metaData, err := GetTokenMetaData(sdkIns, tokenId, contractId)
	if err != nil {
		return 0, 0, lib_error.WrapError(err)
	}
	arr := []uint64{}
	err = json.Unmarshal([]byte(metaData), &arr)
	if err != nil {
		return 0, 0, lib_error.WrapError(err)
	}
	return int(arr[0]), arr[1], nil
}

/**
 * 指定したアカウントが保持しているトークンIdを取得.
 */
func GetTokenIdsByAccount(sdkIns *sdk.IOSTDevSDK, accountName string, contractId string) ([]string, error) {
	// ストレージから所持情報を取得.
	tokenIdsJson, err := GetContractStorageData(sdkIns, contractId, HOLD_TOKEN, accountName)
	if err != nil || len(tokenIdsJson) == 0 {
		return nil, lib_error.WrapError(err)
	}
	tokenIds := []string{}
	err = json.Unmarshal([]byte(tokenIdsJson), &tokenIds)
	if err != nil {
		return nil, lib_error.WrapError(err)
	}
	return tokenIds, nil
}

func GetIsContractSuspended(sdkIns *sdk.IOSTDevSDK, contractId string) (bool, error) {
	r := &rpcpb.GetContractStorageRequest{
		Id:             contractId,
		Key:            "suspended",
		Field:          "",
		ByLongestChain: true,
	}
	res, err := sdkIns.GetContractStorage(r)
	if err != nil {
		return false, err
	}
	suspended, err := strconv.ParseBool(res.Data)
	if err != nil {
		return false, err
	}
	return suspended, nil
}

func SuspendContract(sdkIns *sdk.IOSTDevSDK, contractId string) (string, error) {
	action := sdk.NewAction(contractId, "suspend", "[]")
	txHash, err := sdkIns.SendTxFromActions([]*rpcpb.Action{action})
	return txHash, err
}

func ResumeContract(sdkIns *sdk.IOSTDevSDK, contractId string) (string, error) {
	action := sdk.NewAction(contractId, "resume", "[]")
	txHash, err := sdkIns.SendTxFromActions([]*rpcpb.Action{action})
	return txHash, err
}

func GetTxByHash(sdkIns *sdk.IOSTDevSDK, hash string) (*rpcpb.TransactionResponse, error) {
	txResp, err := sdkIns.GetTxByHash(hash)
	return txResp, err
}

func SetCreatorAddressToTokens(sdkIns *sdk.IOSTDevSDK, contractId, tokenIds, creatorAddress string) (string, error) {
	txHash, err := CallAbi(sdkIns, contractId, "setCreatorAddressToTokens", tokenIds, creatorAddress)
	return txHash, err
}

func RemoveCreatorAddressToTokens(sdkIns *sdk.IOSTDevSDK, contractId, tokenIds string) (string, error) {
	txHash, err := CallAbi(sdkIns, contractId, "removeCreatorAddressFromTokens", tokenIds)
	return txHash, err
}
