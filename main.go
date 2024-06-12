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

	var configPath = flag.String("c", "./config.json", "File name of config(relative or global)")
	flag.Parse()

	// Start ssytem interupt handler
	cs, err := api.NewServerConnection(*configPath)
	if err != nil {
		log.Println("Server connection error:", err.Error())
	}

	// Handling interupt
	interupt := make(chan os.Signal, 1)
	signal.Notify(interupt, os.Interrupt)
	go func() {
		signal := <-interupt
		log.Println("Got interupt signal: ", signal.String())
		cs.Close()
	}()

	// Main loop
	if err := cs.Run(); err != nil {
		log.Println("RUN finished with ERROR:", err)
	}
}
