package api

import "github.com/gorilla/websocket"

type rcError struct {
	err string
}

type ScriptDescriptor struct {
	Name string `json:"name"`
	File string `json:"file"`
}

type Config struct {
	URL      string             `json:"url"`
	Login    string             `json:"login"`
	Password string             `json:"password"`
	Scripts  []ScriptDescriptor `json:"scripts"`
}

type ServerConn struct {
	configFilename string
	config         *Config

	conn       *websocket.Conn
	readerChan chan []byte
	doneChan   chan struct{}
}

// Messages' structs
type registerMsg struct { // register
	Login string `json:"login"`
}

type striptMsg struct { // stript
	Name  string `json:"name"`
	Query string `json:"query"`
}
