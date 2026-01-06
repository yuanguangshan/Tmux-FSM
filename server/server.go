package server

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"tmux-fsm/kernel"
	"tmux-fsm/weaver/adapter"
)

type Config struct {
	SocketPath string
}

type Server struct {
	cfg    Config
	kernel *kernel.Kernel
}

func New(cfg Config) *Server {
	return &Server{
		cfg:    cfg,
		kernel: kernel.New(),
	}
}

func (s *Server) Run(ctx context.Context) error {
	// 清理旧 socket
	_ = os.Remove(s.cfg.SocketPath)

	ln, err := net.Listen("unix", s.cfg.SocketPath)
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Printf("[server] listening on %s\n", s.cfg.SocketPath)

	go s.handleSignals(ctx, ln)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	key, err := protocol.ReadKey(conn)
	if err != nil {
		log.Printf("[server] read error: %v\n", err)
		return
	}

	// ✅ 关键点：Server 不做任何 if / else 判断
	s.kernel.HandleKey(key)
}

func (s *Server) handleSignals(ctx context.Context, ln net.Listener) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case sig := <-ch:
		log.Printf("[server] signal received: %v\n", sig)
	}

	_ = ln.Close()
}
