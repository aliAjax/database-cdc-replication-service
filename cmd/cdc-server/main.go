package main

import (
	"context"
	"flag"
	"github.com/example/cdc-replication/internal/repository"
	"github.com/example/cdc-replication/internal/transport"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	addr := flag.String("addr", ":8087", "HTTP listen address")
	data := flag.String("data", "./var", "state directory")
	flag.Parse()
	store, e := repository.Open(*data)
	if e != nil {
		log.Fatal(e)
	}
	srv := transport.NewServer(store, *addr)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if e := srv.Start(); e != nil {
			log.Printf("server stopped: %v", e)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
	_ = store.Save()
}
