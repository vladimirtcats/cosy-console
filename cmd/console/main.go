package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"cosy-console/config"
	"cosy-console/internal/app"
	"cosy-console/internal/console"
	dbinfra "cosy-console/internal/infrastructure/db"
)

func main() {
	write := flag.Bool("write", false, "allow writes; by default the session is read-only (the server rejects every write)")
	eval := flag.String("e", "", "evaluate a single statement and exit instead of starting the REPL")
	flag.Parse()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("start repl: %v", err)
	}

	var dbOpts []dbinfra.Option
	if !*write {
		dbOpts = append(dbOpts, dbinfra.WithReadOnly())
	}

	container := app.NewContainer(&cfg, dbOpts...)

	// SIGINT belongs to the REPL: it cancels the running statement. If this
	// context listened for SIGINT too, the first Ctrl+C would cancel the
	// exported Ctx for the rest of the session.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()

	if err := console.Run(ctx, container, console.Options{Eval: *eval}); err != nil {
		if errors.Is(err, context.Canceled) {
			return // SIGTERM: clean shutdown
		}

		log.Fatalf("repl exited with error: %v", err)
	}
}
