package main

import (
	"context"
	"os"

	"github.com/urfave/cli/v2"

	accountcmd "github.com/bohdanch-w/go-tgupload/cmd/account"
	configcmd "github.com/bohdanch-w/go-tgupload/cmd/config"
	postcmd "github.com/bohdanch-w/go-tgupload/cmd/post"
	uploadcmd "github.com/bohdanch-w/go-tgupload/cmd/upload"
	versioncmd "github.com/bohdanch-w/go-tgupload/cmd/version"

	whcontext "github.com/bohdanch-w/wheel/context"
	whlogger "github.com/bohdanch-w/wheel/logger"
)

func application(logger whlogger.Logger) *cli.App {
	return &cli.App{
		Name:  "tg-upload",
		Usage: "cli tool to automate uploading to telegra.ph",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "profile",
				EnvVars: []string{"GOTG_PROFILE"},
			},
		},
		Commands: []*cli.Command{
			versioncmd.Version(),
			configcmd.NewCMD(),
			accountcmd.NewCMD(),
			postcmd.NewCMD(logger),
			uploadcmd.NewCMD(logger),
		},
		DefaultCommand: versioncmd.Name,
	}
}

func main() {
	mainLogger := whlogger.NewPtermLogger(whlogger.Info)
	ctx := whcontext.OSInterruptContext(context.Background())

	if err := application(mainLogger).RunContext(ctx, os.Args); err != nil {
		mainLogger.WithError(err).Warnf("command failed")

		os.Exit(1)
	}

	os.Exit(0)
}
