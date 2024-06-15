package api

import (
	"encoding/json"
	"log"
	"net"
	"os/exec"
	"path"

	cw "github.com/CGSG-2021-AE4/go_utils/conn_wrapper"
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
		readerChan: make(chan readMsg, 3),
		doneChan:   make(chan struct{}),
	}, nil
}

func (c *ServerConn) register() error {
	// Write registration
	buf, err := json.Marshal(registerMsg{
		Login: c.config.Login,
	})
	if err != nil {
		return err
	}
	if err := c.conn.Write(cw.MsgTypeRegistration, buf); err != nil {
		return err
	}
	// Wait for response - error/ok
	mt, buf, err := c.conn.Read()
	if err != nil {
		return err
	}
	if mt == cw.MsgTypeClose {
		return rcError("Close msg: " + string(buf))
	}
	if mt != cw.MsgTypeOk || string(buf) != "Registration complete" {
		return rcError("Invalid registration responce: " + string(buf))
	}
	return nil
}

type mainLoopMsg struct {
	Password string          `json:"password"`
	Type     string          `json:"type"`
	Content  json.RawMessage `json:"content"`
}

func (c *ServerConn) handleRequest(buf []byte) error {
	var rawMsg mainLoopMsg
	if err := json.Unmarshal(buf, &rawMsg); err != nil {
		return err
	}
	if rawMsg.Password != c.config.Password {
		return rcError("Wrong password")
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
		return rcError("No script with name: " + msg.Name)
	}
	return rcError("Message type '" + rawMsg.Type + "' is not supported.")
}

func (c *ServerConn) Run() error {
	conn, err := net.Dial("tcp", c.config.URL)
	c.conn = cw.NewConn(conn)
	if err != nil {
		return err
	}
	defer func() {
		c.conn.Conn.Close()
		c.conn = nil
	}()

	if err := c.register(); err != nil {
		return err
	}
	log.Println("REGISTRATION COMPLETE!!!!")

	// Starting reader goroutine
	go func() (err error) {
		defer func() {
			if err != nil {
				log.Println("End reader cycle with error:", err.Error())
			} else {
				log.Println("End reader cycle")
			}
			close(c.doneChan)
		}()

		for {
			mt, buf, err := c.conn.Read()
			if err != nil {
				return err
			}
			if mt == cw.MsgTypeClose {
				log.Println("CLOSE MSG:", string(buf))
				return nil
			}
			c.readerChan <- readMsg{mt, buf}
		}
	}()

	// Reading cycle
	for {
		select {
		case <-c.doneChan:
			return nil
		case msg := <-c.readerChan:
			if msg.mt == cw.MsgTypeRequest {
				if err := c.handleRequest(msg.buf); err != nil {
					if err := c.conn.Write(cw.MsgTypeError, []byte(err.Error())); err != nil {
						return err
					}
				}
			}
		}
	}
}

func (c *ServerConn) Close() error {
	if c.conn == nil {
		return rcError("Socket is not connected")
	}
	log.Println("Closing")
	if err := c.conn.Write(cw.MsgTypeClose, []byte("Buy Buy")); err != nil { // Of course it is not thread safe but now I don't care
		return err
	}
	return nil
}
