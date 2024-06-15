package api

import cw "github.com/CGSG-2021-AE4/go_utils/conn_wrapper"

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

	conn       *cw.ConnWrapper
	readerChan chan readMsg
	doneChan   chan struct{}
}

// Messages' structs
type readMsg struct {
	mt  byte
	buf []byte
}

type registerMsg struct { // register
	Login string `json:"login"`
}

type striptMsg struct { // stript
	Name  string `json:"name"`
	Query string `json:"query"`
}
