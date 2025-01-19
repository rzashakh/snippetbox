package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rzashakh/snippetbox/internal/models"
)

type application struct {
	logger   *slog.Logger
	snippets *models.SnippetModel
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	// Create a new context with a five-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//Database connection logic
	logger.Info("Connecting to database")
	dbpool, err := pgxpool.New(ctx, "postgres://web:pass@localhost:5432/snippetbox")
	if err != nil {
		logger.Error("Unable to create a database connection pool", err)
		os.Exit(1)
	}

	// Test a connection
	if err := dbpool.Ping(ctx); err != nil {
		logger.Error("Unable to ping database", err)
		os.Exit(1)
	}

	defer dbpool.Close()

	app := &application{
		logger:   logger,
		snippets: &models.SnippetModel{DB: dbpool},
	}

	logger.Info("database connection established")
	logger.Info("starting server", "addr", *addr)

	// Call the new app.routes() method to get the servemux containing our routes,
	// and pass that to http.ListenAndServe().
	err = http.ListenAndServe(*addr, app.routes())
	logger.Error(err.Error())
	os.Exit(1)
}
