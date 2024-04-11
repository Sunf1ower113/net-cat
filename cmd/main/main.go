package main

import (
	"fmt"
	"log"
	"net-cat/internal/messenger"
	"os"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	port := messenger.DefaultPort
	args := os.Args[1:]
	if len(args) == 1 {
		port = args[0]
	} else if len(args) > 1 {
		fmt.Println(messenger.Usage)
		os.Exit(0)
	}
	fmt.Printf("Listening on the port %s\n", port)
	server, err := messenger.NewServer("tcp", ":"+port, 10)
	if err != nil {
		log.Fatal(err)
	}
	server.Start()
}
