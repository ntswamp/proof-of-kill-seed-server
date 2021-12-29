package handler

import (
	lib_error "app/src/lib/error"
	"net/http"
)

type WebsocketChatHandler struct {
	ApiBaseHandler
}

func WebsocketChat() HandlerInterface {
	return &WebsocketChatHandler{}
}

func (handler *WebsocketChatHandler) ProcessError(err error) error {
	handler.SetHttpStatus(http.StatusInternalServerError)
	lib_error.WrapError(err)
	return handler.WriteString(err.Error())
}

func (handler *WebsocketChatHandler) Process() error {
	/*
		c := handler.GetContext()
		param := c.Param("roomId")

		go lib_websocket.H.Run()
		lib_websocket.ServeWs(c.Writer, c.Request, param)

		return handler.WriteSuccessJson()
	*/
	return nil
}
