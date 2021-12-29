package lib_puzzle

import "testing"

func TestParseText(t *testing.T) {
	oktext := "123456789123456789123456789123456789123456089123456789123456789123456789123456789"
	result := parseText(oktext)

	//row 5 col 7, that's 0
	if result[4][6] != 0 {
		t.Error("want: 0, got:", result[4][6])
	}
}

func TestIsBoardSolvable(t *testing.T) {
	//solution is 735614892842973561961285374286349157413857926579126438157492683694738215328561749
	bad := "735614892842973561961285374286349157413857926579126438157492683694738215328561790"
	//bad = "111111111111111111111111111111111111111111111111111111111111111111111111111111111"
	//bad = "000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	good := "735614892842973561961285374286349157413857926579126438157492683694738215328561740"

	if IsBoardSolvable(bad) != false || IsBoardSolvable(good) != true {
		t.Error("want: false true got:", IsBoardSolvable(bad), IsBoardSolvable(good))
	}

}