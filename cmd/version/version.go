package version

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/bohdanch-w/go-tgupload/internal/build"
)

const (
	Name = "version"
)

func Version() *cli.Command {
	return &cli.Command{
		Name:   Name,
		Usage:  "get gotg basic info",
		Action: version(),
	}
}

func version() cli.ActionFunc {
	return func(cCtx *cli.Context) error {
		fmt.Fprintln(os.Stdout, "GoTg:")
		fmt.Fprintf(os.Stdout, "  version: %s\n", build.Version)
		fmt.Fprintf(os.Stdout, "  go:      %s\n", build.GoVersion)
		fmt.Fprintf(os.Stdout, "  at:      %s\n", build.BuiltAt)

		return nil
	}
}
