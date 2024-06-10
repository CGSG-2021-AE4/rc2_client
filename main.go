package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

var login = "AAAAA"

// Messages' structs
type registerMsg struct { // register
	Login string `json:"login"`
}

func main() {
	fmt.Println("CGSG forever!!!")

	interupt := make(chan os.Signal, 1)
	signal.Notify(interupt, os.Interrupt)

	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:3047/client_service", nil)
	if err != nil {
		log.Fatal("Dial error: ", err)
	}
	defer c.Close()

	// Write registration
	buf, err := WriteMsg("registration", registerMsg{
		Login: login,
	})
	if err != nil {
		log.Println("Registration json error: ", err)
		return
	}
	if err := c.WriteMessage(websocket.BinaryMessage, buf); err != nil {
		log.Println("Registration write error: ", err)
		return
	}
	_, buf, err = c.ReadMessage()
	mt, msg, err := ReadMsg[string](buf)
	if mt == "error" {
		log.Println("Registration error: ", msg)
		return
	} else if mt != "msg" || msg != "Registration complete" {
		log.Println("Invalid registration responce: ", msg)
		return
	}
	log.Println("REGISTRATION COMPLETE!!!!")

	for {
		wsmt, buf, err := c.ReadMessage()
		if err != nil {
			log.Println("READ ERROR: ", err)
			return
		}
		log.Println("GO MSG: ", websocket.FormatMessageType(wsmt), buf)
	}

}
