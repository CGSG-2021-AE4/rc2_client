package api

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type rcError struct {
	err string
}

func (e rcError) Error() string {
	return e.err
}

func NewError(msg string) rcError {
	return rcError{
		err: msg,
	}
}

// Messages' structs
type registerMsg struct { // register
	Login string `json:"login"`
}

type ServerConn struct {
	serverURL  string
	login      string
	password   string
	conn       *websocket.Conn
	readerChan chan []byte
	doneChan   chan struct{}
}

func NewServerConnection(url string, login string, password string) *ServerConn {
	return &ServerConn{
		serverURL:  url,
		login:      login,
		password:   password,
		conn:       nil,
		readerChan: make(chan []byte, 3),
		doneChan:   make(chan struct{}),
	}
}

func (c *ServerConn) register() error {
	// Write registration
	buf, err := WriteMsg("registration", registerMsg{
		Login: c.login,
	})
	if err != nil {
		return err
	}
	if err := c.conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
		return err
	}
	wsmt, buf, err := c.conn.ReadMessage()
	if err != nil {
		return err
	}
	if wsmt == websocket.CloseMessage {
		return NewError("Close msg: " + string(buf))
	}
	mt, msg, err := ReadMsg[string](buf)
	if err != nil {
		return err
	}
	if mt != "msg" || msg != "Registration complete" {
		return NewError("Invalid registration responce: " + msg)
	}
	return nil
}

func (c *ServerConn) Run() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.serverURL, nil)
	c.conn = conn
	if err != nil {
		return err
	}
	defer func() {
		c.conn.Close()
		c.conn = nil
	}()

	if err := c.register(); err != nil {
		return err
	}
	log.Println("REGISTRATION COMPLETE!!!!")

	// Start reader

	// Starting reader goroutine
	go func() {
		for {
			wsmt, buf, err := c.conn.ReadMessage()
			if err != nil {
				fmt.Println("READ ERROR:", err)
				break
			}
			if wsmt == websocket.CloseMessage {
				log.Println("CLOSE MSG:", string(buf))
				break
			} else if wsmt == websocket.TextMessage {
				log.Println("TEXT MSG:", string(buf))
			} else if wsmt == websocket.BinaryMessage {
				c.readerChan <- buf
			}
		}
		log.Println("Close reader goroutine")
		close(c.doneChan)
	}()

	for {
		select {
		case <-c.doneChan:
			break
		case buf := <-c.readerChan:
			log.Println("GOT BIN MSG:", string(buf))
		}
	}
	log.Print("Close read cycle")
	return nil
}

func (c *ServerConn) Close() error {
	if c.conn == nil {
		return NewError("Socket is not connected")
	}
	if err := c.conn.WriteMessage(websocket.CloseMessage, []byte("Buy Buy")); err != nil { // Of course it is not thread safe but now I don't care
		return err
	}
	return nil
}
