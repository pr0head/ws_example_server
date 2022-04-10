package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

const (
	MessageSetServerStatus = "set_server_status"
	MessageAddGameChar     = "add_game_char"
	MessageSendGameBalance = "send_game_balance"
	MessageGetGameBalance  = "get_game_balance"
)

var (
	supportedMessages = map[string]bool{
		MessageSendGameBalance: true,
		MessageSetServerStatus: true,
		MessageAddGameChar:     true,
		MessageGetGameBalance:  true,
	}
)

type wbs struct {
	mu         sync.Mutex
	conn       *websocket.Conn
	pingPeriod time.Duration
	pongWait   time.Duration
	writeWait  time.Duration
	send       chan []byte
}

func NewWebSocket(ws *websocket.Conn, pongWait, writeWait time.Duration) *wbs {
	w := &wbs{
		conn:       ws,
		pingPeriod: (pongWait * 9) / 10,
		pongWait:   pongWait,
		writeWait:  writeWait,
		send:       make(chan []byte),
	}
	go w.run()

	return w
}

func (w *wbs) run() {
	pingTicker := time.NewTicker(w.pingPeriod)

	defer func() {
		pingTicker.Stop()
		w.conn.Close()
	}()

	for {
		select {
		case data, ok := <-w.send:
			w.conn.SetWriteDeadline(time.Now().Add(w.writeWait))

			if !ok {
				w.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := w.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println("write message error:", err)
				return
			}
		case <-pingTicker.C:
			log.Println("push ping")
			w.conn.SetWriteDeadline(time.Now().Add(w.writeWait))
			if err := w.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (w *wbs) SendGetGameBalance() error {
	log.Println("send: ", MessageGetGameBalance)
	data := &GetGameBalance{
		UserId:     time.Now().String(),
		ServerName: "server_name",
	}

	if err := w.writeTextMessage(MessageGetGameBalance, data); err != nil {
		return err
	}

	return nil
}

func (w *wbs) SendSendGameBalance() error {
	log.Println("send: ", MessageSendGameBalance)
	data := &SendGameBalance{
		UserId:     time.Now().String(),
		ServerName: "server_name",
		Tokens: []*GameTokens{
			{
				Id:     "token_id",
				Amount: 3.14,
			},
		},
	}

	if err := w.writeTextMessage(MessageSendGameBalance, data); err != nil {
		return err
	}

	return nil
}

func (w *wbs) SendAddGameChar() error {
	log.Println("send: ", MessageAddGameChar)
	data := &AddGameChar{
		UserId:     time.Now().String(),
		ServerName: "server_name",
		CharId:     "char_id",
		CharName:   "char_name",
	}

	if err := w.writeTextMessage(MessageAddGameChar, data); err != nil {
		return err
	}

	return nil
}

func (w *wbs) SendSetServerStatus() error {
	log.Println("send: ", MessageSetServerStatus)
	data := &SetServerStatus{
		Name:     time.Now().String(),
		IsActive: true,
	}

	if err := w.writeTextMessage(MessageSetServerStatus, data); err != nil {
		return err
	}

	return nil
}

func (w *wbs) Listen() {
	defer w.conn.Close()

	w.conn.SetReadDeadline(time.Now().Add(w.pongWait))
	w.conn.SetPongHandler(func(string) error {
		w.conn.SetReadDeadline(time.Now().Add(w.pongWait))
		return nil
	})

	for {
		mt, msg, err := w.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		switch mt {
		case websocket.CloseMessage:
			log.Println("receive close message")
		case websocket.TextMessage:
			if err := w.parseMessage(msg); err != nil {
				log.Println("unable to parse message:", err)
				return
			}
		default:
			log.Println("receive unsupported message type", msg)
		}
	}
}

func (w *wbs) parseMessage(message []byte) error {
	payload := &WebSocketMessage{}

	if err := json.Unmarshal(message, payload); err != nil {
		return err
	}

	if _, ok := supportedMessages[payload.Type]; !ok {
		return fmt.Errorf("unsupported payload type: %s", payload.Type)
	}

	log.Println("receive: ", payload.Type, payload.Data)

	return nil
}

func (w *wbs) writeTextMessage(t string, data interface{}) error {
	payload, err := json.Marshal(&WebSocketMessage{
		Type: t,
		Data: data,
	})

	if err != nil {
		log.Println("message marshal:", err)
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.send <- payload

	return nil
}
