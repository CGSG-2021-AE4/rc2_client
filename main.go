package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"ws_client/api"
)

func main() {
	fmt.Println("CGSG forever!!!")

	var login = flag.String("l", "Default", "Login")
	var password = flag.String("p", "12345", "Password")
	flag.Parse()

	fmt.Println("Login:", *login, "Password:", *password)

	interupt := make(chan os.Signal, 1)
	signal.Notify(interupt, os.Interrupt)
	// Start ssytem interupt handler
	cs := api.NewServerConnection("ws://localhost:3047/client_service", *login, *password)
	go func() {
		signal := <-interupt
		log.Println("Got interupt signal: ", signal.String())
		cs.Close()
	}()
	if err := cs.Run(); err != nil {
		log.Println("RUN finished with ERROR:", err)
	}
}
