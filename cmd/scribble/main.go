package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/SladkyCitron/slogcolor"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nasermirzaei89/scribble"
)

func main() {
	ctx := context.Background()

	{
		opts := slogcolor.DefaultOptions
		opts.Level = scribble.GetLogLevelFromEnv()
		slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, opts)))
	}

	err := run(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to run", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	app, err := scribble.NewApp(ctx)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	err = app.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run app: %w", err)
	}

	return nil
}
