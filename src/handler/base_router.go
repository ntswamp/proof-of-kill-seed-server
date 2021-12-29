package handler

import (
	"html/template"
	"net/http"

	"app/src/constant"
	lib_basic_auth "app/src/lib/auth"
	lib_websocket "app/src/lib/websocket"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func GETPOST(group *gin.RouterGroup, path string, funcs ...gin.HandlerFunc) {
	group.GET(path, funcs...)
	group.POST(path, funcs...)
}

func addHtml(renderer multitemplate.Renderer, f func() (string, []string, template.FuncMap)) {
	name, htmls, funcMap := f()
	if funcMap == nil {
		funcMap = template.FuncMap{}
	}
	renderer.AddFromFilesFuncs(name, funcMap, htmls...)
}

func RouteApi(engine *gin.Engine) {
	if constant.IsManagement() {
		return
	}

	// Session Store
	sessStore := cookie.NewStore([]byte(constant.SESSION_AUTHKEY), []byte(constant.SESSION_CIPHERKEY))
	sessOptions := sessions.Options{
		Path:     "/",
		Domain:   constant.GetTokenLinkDomain(),
		MaxAge:   0,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	if constant.IsProduction() {
		sessOptions.Secure = true
	} else {
		sessOptions.Secure = false
	}

	sessStore.Options(sessOptions)
	engine.Use(sessions.Sessions(constant.SESSION_NAME, sessStore))
	renderer := multitemplate.NewRenderer()
	engine.LoadHTMLGlob("../html/chatroom_demo.html")

	/**
	*
	*	SYSTEMATIC UTILITY
	*
	**/
	engine.GET("/health", Wrap(HealthCheck))
	engine.GET("/idol/:idolId/:userId/:language", Wrap(Chatroom))
	addHtml(renderer, ChatroomHtml)
	engine.GET("/ws/:idolId/:userId/:language", func(c *gin.Context) {
		go lib_websocket.H.Run()
		lib_websocket.ServeWs(c.Writer, c.Request, c)
	})

	//alternative
	//engine.GET("/ws/:idolId", Wrap(WebsocketChat))
	/**
	*
	*	ACCOUNT SYSTEM
	*
	**/
	engine.POST("/account/name/set", Wrap(NameSet))
	engine.POST("/account/name/get", Wrap(NameGet))

	// Static
	if !constant.IsProduction() {
		staticGroup := engine.Group("/static", lib_basic_auth.Auth())
		staticGroup.Static("/", "static")
	}

	engine.OPTIONS("/:a", func(context *gin.Context) {
		context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		context.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		context.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		context.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		context.String(204, "")
	})

}
