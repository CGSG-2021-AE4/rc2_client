package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"path"

	"github.com/gorilla/websocket"
)

func NewServerConnection(configFilename string) (*ServerConn, error) {
	config, err := LoadConfig(configFilename)
	if err != nil {
		return nil, err
	}

	return &ServerConn{
		configFilename: configFilename,
		config:         config,

		conn:       nil,
		readerChan: make(chan []byte, 3),
		doneChan:   make(chan struct{}),
	}, nil
}

func (c *ServerConn) register() error {
	// Write registration
	buf, err := WriteMsg("registration", registerMsg{
		Login: c.config.Login,
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

type mainLoopMsg struct {
	Password string          `json:"password"`
	Type     string          `json:"type"`
	Content  json.RawMessage `json:"content"`
}

func (c *ServerConn) handleMsg(buf []byte) error {
	var rawMsg mainLoopMsg
	if err := json.Unmarshal(buf, &rawMsg); err != nil {
		return err
	}
	if rawMsg.Password != c.config.Password {
		return NewError("Wrong password")
	}
	switch rawMsg.Type {
	case "script":
		var msg striptMsg
		if err := json.Unmarshal(rawMsg.Content, &msg); err != nil {
			return err
		}
		for i := range len(c.config.Scripts) {
			if c.config.Scripts[i].Name == msg.Name {
				filepath := c.config.Scripts[i].File
				if !path.IsAbs(filepath) {
					filepath = path.Join(path.Dir(c.configFilename), filepath)
				}
				cmd := exec.Command(filepath + " " + msg.Query)
				if err := cmd.Start(); err != nil {
					return err
				}
				return nil
			}
		}
		return NewError("No script with name: " + msg.Name)
	}
	return NewError("Message type '" + rawMsg.Type + "' is not supported.")
}

func (c *ServerConn) Run() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.config.URL, nil)
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
		close(c.doneChan)
	}()

	// Reading cycle
	for {
		select {
		case <-c.doneChan:
			return nil
		case buf := <-c.readerChan:
			msg := "OK"
			if err := c.handleMsg(buf); err != nil {
				msg = "ERROR: " + err.Error()
				log.Println("HANDLE ERROR:", err.Error())
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, []byte(msg)); err != nil {
				log.Println("WRITE ERROR: ", err.Error())
			}
		}
	}
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
