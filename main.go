package main

import (
	"context"
	"flag"
	"log"

	"Tmux-FSM/server"
)

func main() {
	serverMode := flag.Bool("server", false, "run as server")
	socketPath := flag.String("socket", "/tmp/tmux-fsm.sock", "socket path")
	flag.Parse()

	if *serverMode {
		srv := server.New(server.Config{
			SocketPath: *socketPath,
		})
		log.Fatal(srv.Run(context.Background()))
		return
	}

	// client / other modes 保持你原来的逻辑
	log.Println("no mode specified")
}
