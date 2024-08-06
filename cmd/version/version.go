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
			os.Stdout.WriteString(fmt.Sprintf("tg-upload version:    %s\n", build.Version))
			os.Stdout.WriteString(fmt.Sprintf("tg-upload build date: %s\n", build.Date))

			return nil
		},
	}
}
