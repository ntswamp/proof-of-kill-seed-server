package lib_sig

import (
	"app/src/constant"
	"testing"
)

func TestGenerateSignatureFromBoard(t *testing.T) {
	//goodBoard := "735614892842973561961285374286349157413857926579126438157492683694738215328561740"
	t1 := "answer_puzzlepe_test_01PL_NP1"
	t2 := "create_puzzlehakumaiPL_NP020004001500090004060701009040900100702065900006200000008000600094008000000640300"
	t3 := "buy_puzzlepe_test_00PL_NP1"
	text := []string{t1, t2, t3}

	s1 := "5zavLBCejcr7DiFeebQ2H1NHzxDWF6fHEryfS1F5jj1UkHvnRakjpPRDfjtrquK6APCZLpdLLVEaWL8U2Fy376QF"
	s2 := "8cNQGEGXjaH4hHXVqtij7SFuducbQzksLQ6fko7EX4A4oPbrnhvLfSSufMnKKcWTGdrs7cGuMyEHC8FNXbhNYoz"
	s3 := "2UYhuaFF4iarhkQ7bqm7Fyz14jv5yzFtM8HnoTVvHXjV4JGXdsGTzMHy17PhebgUkLQMRSS3piof4Juc879oDzFP"

	signature := []string{s1, s2, s3}

	for i, text := range text {
		sig, err := GenerateSignatureFromText(text, constant.ADMIN_IOST_ACCOUNT_PRIVATE_KEY)
		if err != nil {
			t.Fatal(err)
		}
		if sig != signature[i] {
			t.Fatalf("want:%s, got:%s\n", signature[i], sig)
		}
	}
}

func TestIsLegalMetamaskSig(t *testing.T) {
	legalSig := IsLegalSig(
		"0x4eE96fE0e4B1975d0AFf76e12D7BB3c765394342",
		"0xb1576c0135c1753e269f0154b2a84e09a687c3cbfb2d9b086dfe1271e4fccf075538ffd968258daecf92c5bc56b9a651a52112a09f19804166df3eaf2661a3ac1b",
		[]byte("hello"),
	)

	if !legalSig {
		t.Fatalf("want: ture, got:%v\n", legalSig)
	}
}
