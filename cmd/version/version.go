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
		Name: Name,
		Action: func(_ *cli.Context) error {
			os.Stdout.WriteString(fmt.Sprintf("backoffice version:    %s\n", build.Version))
			os.Stdout.WriteString(fmt.Sprintf("backoffice build:      %s\n", build.Build))
			os.Stdout.WriteString(fmt.Sprintf("backoffice build date: %s\n", build.Date))

			return nil
		},
	}
}
