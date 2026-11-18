package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antoniocra04/racion/internal/store"
	"github.com/antoniocra04/racion/internal/web"
)

func main() {
	address := flag.String("addr", "127.0.0.1:8080", "HTTP listen address")
	database := flag.String("db", "data/ration.db", "SQLite database path")
	backup := flag.String("backup", "", "create a SQLite backup at this new path and exit")
	flag.Parse()
	if err := run(*address, *database, *backup); err != nil {
		log.Fatal(err)
	}
}

func run(address, database, backup string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if backup != "" {
		if _, err := os.Stat(database); err != nil {
			return fmt.Errorf("open backup source: %w", err)
		}
	}
	db, err := store.Open(ctx, database)
	if err != nil {
		return err
	}
	defer db.Close()
	if backup != "" {
		if err := db.Backup(ctx, backup); err != nil {
			return fmt.Errorf("backup database: %w", err)
		}
		log.Printf("Backup saved to %s", backup)
		return nil
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler:           web.New(db),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.Serve(listener)
	}()
	log.Printf("Рацион · http://%s", listener.Addr())
	select {
	case err := <-serverDone:
		return err
	case <-ctx.Done():
		shutdown, finish := context.WithTimeout(context.Background(), 5*time.Second)
		defer finish()
		if err := server.Shutdown(shutdown); err != nil {
			return err
		}
		err := <-serverDone
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
