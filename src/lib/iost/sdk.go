package lib_iost

import (
	"app/src/constant"
	lib_error "app/src/lib/error"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/crypto"
	rpcpb "github.com/iost-official/go-iost/rpc/pb"
	"github.com/iost-official/go-iost/sdk"
)

type IOSTDevSDKOptions struct {
	AccName string
	SecKey  string
	Server  string
	ChainId uint32
	TxInfo  *IOSTDevSDKTxInfo
}

type IOSTDevSDKTxInfo struct {
	GasLimit    float64
	GasRatio    float64
	Expiration  int64
	DelaySecond int64
	AmountLimit []*rpcpb.AmountLimit
}

/**
 * go-iostのsdkのインスタンス作成.
 * サーバとアカウント設定も行う.
 */
func NewIOSTDevSDK() (*sdk.IOSTDevSDK, error) {
	sdkIns := sdk.NewIOSTDevSDK()
	err := setupIOSTDevSDK(sdkIns)
	if err != nil {
		return nil, lib_error.WrapError(err)
	}
	return sdkIns, nil
}

func NewIOSTDevSDKWithOptions(opt *IOSTDevSDKOptions) (*sdk.IOSTDevSDK, error) {
	sdkIns := sdk.NewIOSTDevSDK()
	err := setupIOSTDevSDKWithOptions(sdkIns, opt)
	if err != nil {
		return nil, lib_error.WrapError(err)
	}
	return sdkIns, nil
}

func setupIOSTDevSDKWithOptions(sdkIns *sdk.IOSTDevSDK, opt *IOSTDevSDKOptions) error {
	// Server and chain id
	if opt.Server != "" && opt.ChainId != 0 {
		sdkIns.SetServer(opt.Server)
		sdkIns.SetChainID(opt.ChainId)
	} else {
		if constant.IsDebug() {
			// Testnet
			sdkIns.SetServer("13.52.105.102:30002")
			sdkIns.SetChainID(1023)
		} else {
			// Mainnet
			sdkIns.SetServer("54.180.196.80:30002")
			sdkIns.SetChainID(1024)
		}
	}
	// Gas limit
	if opt.TxInfo != nil {
		sdkIns.SetTxInfo(opt.TxInfo.GasLimit, opt.TxInfo.GasRatio, opt.TxInfo.Expiration, opt.TxInfo.DelaySecond, opt.TxInfo.AmountLimit)
	}
	// Key pair creation
	seckeyByte := common.Base58Decode(opt.SecKey)
	kp, err := account.NewKeyPair(seckeyByte, crypto.NewAlgorithm(""))
	if err != nil {
		return lib_error.WrapError(err)
	}
	// Set Acc
	sdkIns.SetAccount(opt.AccName, kp)
	return nil
}

/**
 * IOSTDevSDKをセットアップ.
 * サーバとアカウント設定を行う.
 */
func setupIOSTDevSDK(sdkIns *sdk.IOSTDevSDK) error {
	// サーバ設定.
	if constant.IsDebug() {
		// testnet.
		sdkIns.SetServer("13.52.105.102:30002")
		sdkIns.SetChainID(1023)
	} else {
		// mainnet.
		sdkIns.SetServer("54.180.196.80:30002")
		sdkIns.SetChainID(1024)
	}
	// キーペアの作成.
	secKey := GetContractAccountKey()
	seckeyByte := common.Base58Decode(secKey)
	kp, err := account.NewKeyPair(seckeyByte, crypto.NewAlgorithm(""))
	if err != nil {
		return lib_error.WrapError(err)
	}
	// アカウント設定.
	account := GetContractAccountName()
	sdkIns.SetAccount(account, kp)
	// TX info settings
	sdkIns.SetTxInfo(2000000, 1.0, 90, 0, []*rpcpb.AmountLimit{{Token: "*", Value: "unlimited"}})
	return nil
}

/**
 * 現在の不可逆ブロック番号を取得する.
 */
func GetLatestIrreversibleBlock(sdk *sdk.IOSTDevSDK) (int64, error) {
	// ブロックチェーン情報.
	info, err := sdk.GetChainInfo()
	if err != nil {
		return 0, lib_error.WrapError(err)
	}
	return info.LibBlock, nil
}
