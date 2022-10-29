package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	postcmd "github.com/bohdanch-w/go-tgupload/cmd/post"
	versioncmd "github.com/bohdanch-w/go-tgupload/cmd/version"
	"github.com/bohdanch-w/go-tgupload/helpers"
)

func application(logger *zap.Logger) *cli.App {
	return &cli.App{
		Name:  "uploader",
		Usage: "cli tool to automate uploading to telegra.ph",
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			versioncmd.Version(),
			postcmd.NewCMD(logger),
		},
	}
}

func _main() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			return
		case <-shutdown:
			cancel()
		}
	}()

	mainLogger := helpers.MustLogger()
	defer func() { _ = mainLogger.Sync() }()

	if err := application(mainLogger).RunContext(ctx, os.Args); err != nil {
		mainLogger.Error("command failed", zap.Error(err))

		return 1
	}

	return 0
}

func main() {
	os.Exit(_main())
}
