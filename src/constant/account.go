package constant

//account type
const (
	USER  = 0
	ADMIN = 1
)

var PERMISSION_TEXT = map[int]string{
	0: "User",
	1: "Admin",
}

//language
const (
	JAPANESE = 0
	CHINESE  = 1
	ENGLISH  = 2
)
