package lib_websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	lib_ai "app/src/lib/ai"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	WRITE_WAIT = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	PONG_WAIT = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	PING_PERIOD = (PONG_WAIT * 9) / 10

	// Maximum message size allowed from peer.
	MAX_MSG_SIZE = 512

	//AI's identifier ID
	AI_ID = "idolverse-ai-identifier"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	// Registered connections.
	rooms map[string]map[*connection]bool

	//3 types of events.
	broadcast chan message // Inbound messages from the connections.

	register chan subscription // Register requests from the connections.

	unregister chan subscription // Unregister requests from connections.
}

var H = Hub{
	broadcast:  make(chan message),
	register:   make(chan subscription),
	unregister: make(chan subscription),
	rooms:      make(map[string]map[*connection]bool),
}

type message struct {
	data []byte
	room string
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

type subscription struct {
	conn *connection
	room string
}

func (hub *Hub) Run() {
	for {
		select {
		case s := <-hub.register:
			connections := hub.rooms[s.room]
			if connections == nil {
				connections = make(map[*connection]bool)
				hub.rooms[s.room] = connections
			}
			hub.rooms[s.room][s.conn] = true
		case s := <-hub.unregister:
			connections := hub.rooms[s.room]
			if connections != nil {
				if _, ok := connections[s.conn]; ok {
					delete(connections, s.conn)
					close(s.conn.send)
					if len(connections) == 0 {
						delete(hub.rooms, s.room)
					}
				}
			}
		case message := <-hub.broadcast:
			connections := hub.rooms[message.room]
			for c := range connections {
				select {
				case c.send <- message.data:
				default:
					close(c.send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(hub.rooms, message.room)
					}
				}
			}
		}
	}
}

// readUserInput pumps messages from the websocket connection to the hub.
func (s subscription) readUserInput(userId string, language string) {
	c := s.conn
	defer func() {
		H.unregister <- s
		c.ws.Close()
	}()
	c.ws.SetReadLimit(MAX_MSG_SIZE)
	c.ws.SetReadDeadline(time.Now().Add(PONG_WAIT))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(PONG_WAIT)); return nil })

	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		// display/hide self-input message
		//m := message{msg, s.room}
		//H.broadcast <- m

		//AI respond
		//TODO: language recognition
		res := lib_ai.ProduceBinaryResponse(userId, string(msg), language)

		aiMsg := message{[]byte(res), s.room}
		H.broadcast <- aiMsg

	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(messageType int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
	return c.ws.WriteMessage(messageType, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (s *subscription) writePump() {
	c := s.conn
	ticker := time.NewTicker(PING_PERIOD)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (s *subscription) AiGreeting(idolName string, userId string, lang string) {

	res := lib_ai.ProduceBinaryResponse(userId, "こんにちわ", lang)

	var result map[string]string
	json.Unmarshal(res, &result)

	for key, msg := range result {
		result[key] = userId + ", " + msg
	}
	response, _ := json.Marshal(result)

	aiMsg := message{response, s.room}
	H.broadcast <- aiMsg
}

// serveWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request, context *gin.Context) {
	idolId, _ := context.Params.Get("idolId")
	userId, _ := context.Params.Get("userId")
	language, _ := context.Params.Get("language")

	if language != "en" && language != "zh" && language != "ja" {
		log.Println(errors.New("unknown language"))
		return
	}

	fmt.Print(idolId)
	log.Println(r.Header)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}
	s := subscription{c, idolId}
	H.register <- s
	//TODO: language setting
	s.AiGreeting(idolId, userId, language)
	go s.writePump()
	go s.readUserInput(userId, language)
}
