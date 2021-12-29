package constant

import (
	"os"
)

// 環境名.
var SERVER_TYPE string = os.Getenv("SERVER_TYPE")

// プロジェクトのroot.
const PROJECT_ROOT = "/home/ec2-user/golang/unity-idolverse-server"

// ログの保存先.
const LOG_DIR = "/home/ec2-user/golang/unity-idolverse-server/log/"

// テンプレートがあるディレクトリ.
const TEMPLATE_DIR = "/home/ec2-user/golang/unity-idolverse-server/template/"

// アプリコードネーム.
const APP_CODENAME = "idolverse"

// 開発者IPアドレス.
var DEV_IP = map[string]bool{
	"211.9.52.235": true, // プラチナエッグ東京.
}

func GetServerEnv() string {
	return SERVER_TYPE
}

func IsLocal() bool {
	switch GetServerEnv() {
	case "local":
		return true
	}
	return false
}

func IsDebug() bool {
	name := GetServerEnv()
	switch name {
	case "production", "management":
		return false
	}
	return true
}

func IsProduction() bool {
	switch GetServerEnv() {
	case "production", "management":
		return true
	}
	return false
}

func IsManagement() bool {
	switch GetServerEnv() {
	case "management":
		return true
	}
	return false
}

// クライアント側で操作するs3のバケット名.
func GetClientS3BucketName() string {
	switch GetServerEnv() {
	case "production":
		return "idolverse-prod"
	default:
		return "idolverse-dev"
	}
}

func GetStaticS3BucketName() string {
	switch GetServerEnv() {
	case "production", "management":
		return "idolverse-static-prod"
	}
	return "idolverse-static-dev"
}

// Sale Site URL
func GetSaleSiteURL() string {
	switch GetServerEnv() {
	case "local":
		return "http://localhost:9090/static"
	case "development":
		return "http://ec2-54-249-155-208.ap-northeast-1.compute.amazonaws.com/static"
	case "production":
		return "https://www.tokenlink.io"
	}
	return ""
}

func GetTokenLinkDomain() string {
	switch GetServerEnv() {
	case "development":
		return "ec2-54-249-155-208.ap-northeast-1.compute.amazonaws.com"
	case "production":
		return "tokenlink.io"
	}
	return "localhost"
}
