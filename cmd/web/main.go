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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Info("Connecting to database")
	dbpool, err := pgxpool.New(ctx, "postgres://web:pass@localhost:5432/snippetbox")
	if err != nil {
		logger.Error("Unable to create a database connection pool", err)
		os.Exit(1)
	}

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

	err = http.ListenAndServe(*addr, app.routes())
	logger.Error(err.Error())
	os.Exit(1)
}
