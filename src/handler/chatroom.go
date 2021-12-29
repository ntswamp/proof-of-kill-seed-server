package handler

import (
	lib_error "app/src/lib/error"
	lib_websocket "app/src/lib/websocket"
	"html/template"
	"net/http"

	"github.com/go-errors/errors"
)

type ChatroomHandler struct {
	BaseHandler
}

func (self *ChatroomHandler) ProcessError(err error) error {
	self.SetHttpStatus(http.StatusInternalServerError)
	lib_error.WrapError(err)
	return self.WriteString(errors.Wrap(err, 1).ErrorStack())
}

func (self *ChatroomHandler) Process() error {

	go lib_websocket.H.Run()
	return self.WriteHtml("chatroom_demo.html")
}

func Chatroom() HandlerInterface {
	return &ChatroomHandler{}
}

func ChatroomHtml() (string, []string, template.FuncMap) {
	funcMap := template.FuncMap{}
	return "chatroom_demo.html", []string{"../html/chatroom_demo.html"}, funcMap
}
