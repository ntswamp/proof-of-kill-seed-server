package lib_basic_auth

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type authPair struct {
	authValue string
	user      string
}

var accounts = []authPair{
	{
		authValue: "sale_test:UtFZKSSu6W",
		user:      "sale_test",
	},
}

var authorizedIps = map[string]bool{
	"60.86.149.117": true,
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check auth for ips not in authorizedIps map
		if ok := authorizedIps[c.ClientIP()]; !ok {
			authHeader := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)
			if authHeader[0] != "Basic" || len(authHeader) != 2 {
				setUnauthorized(c)
				return
			}
			authValue, err := base64.StdEncoding.DecodeString(authHeader[1])
			if err != nil {
				setUnauthorized(c)
				return
			}
			found, user := searchAccounts(string(authValue))
			if !found {
				setUnauthorized(c)
				return
			}
			c.Set("user", user)
		}
		c.Next()
	}
}

func searchAccounts(authValue string) (bool, string) {
	for _, acc := range accounts {
		if acc.authValue == authValue {
			return true, acc.user
		}
	}
	return false, ""
}

func setUnauthorized(c *gin.Context) {
	c.Header("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
	c.AbortWithStatus(http.StatusUnauthorized)
}
