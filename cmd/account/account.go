package account

import (
	"fmt"
	"os"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/bohdanch-w/go-tgupload/config"
	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/pkg/utils"
	"github.com/bohdanch-w/wheel/ds/hashset"
	wherr "github.com/bohdanch-w/wheel/errors"
)

const (
	Name = "account"
)

func NewCMD() *cli.Command {
	return &cli.Command{
		Name:  Name,
		Usage: "manage account",
		Subcommands: []*cli.Command{
			{
				Name:   "login",
				Action: login,
			},
			{
				Name:   "setup",
				Action: setup,
			},
			{
				Name:   "token",
				Action: showToken,
			},
		},
		Action: show,
	}
}

func setup(ctx *cli.Context) error {
	profile, err := setUpSelectProfile(ctx)
	if err != nil {
		return err
	}

	cfg, err := config.ReadConfig(profile)
	if err != nil {
		return err
	}

	acc := cfg.Account()
	if err := promtUserAccount(&acc); err != nil {
		return err
	}

	prompt := promptui.Prompt{
		Label: fmt.Sprintf(
			`Confirm changes to %q profile
 - Name:       %s
 - Short name: %s
 - Author URL: %s`,
			cfg.Profile,
			acc.AuthorName,
			acc.AuthorShortName,
			acc.AuthorURL,
		),
		IsConfirm: true,
	}

	confirmed, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("get confirmation: %w", err)
	}

	if confirmed != "Y" {
		return nil
	}

	cfg.SetAccount(acc)

	if err := config.StoreConfig(cfg); err != nil {
		return fmt.Errorf("store config: %w", err)
	}

	return nil
}

func setUpSelectProfile(ctx *cli.Context) (string, error) {
	// specific profile provided
	profile := ctx.String("profile")
	if profile != "" {
		return profile, nil
	}

	// select from existing profiles
	profiles, err := config.ListProfiles()
	if err != nil {
		return "", fmt.Errorf("find available profiles: %w", err)
	}

	profiles = append(profiles, "+ new")

	// // profileListPrompt := promptui.Select{
	// // 	Label: "Select profile",
	// // 	Items: profiles,
	// // }

	// // idx, selectedProfile, err := profileListPrompt.Run()
	// // if err != nil {
	// // 	return "", fmt.Errorf("select profile: %w", err)
	// // }

	// if idx < len(profiles)-1 {
	// 	return selectedProfile, nil
	// }

	selectedProfile := prompter.Choose("Select profile", profiles, profiles[0])
	if selectedProfile != "+ new" {
		return selectedProfile, nil
	}

	// create new profile
	// newProfilePrompt := promptui.Prompt{
	// 	Label: "Name new profile",
	// 	Validate: func(input string) error {
	// 		if pfs.Has(input) {
	// 			return wherr.Error("profile already exists")
	// 		}

	// 		return nil
	// 	},
	// }

	// name, err := newProfilePrompt.Run()
	// if err != nil {
	// 	return "", fmt.Errorf("name new profile: %w", err)
	// }
	for {
		name := prompter.Prompt("Profile ID", "")
		if name == "" {
			continue
		}

		pfs := hashset.New(profiles...)
		if pfs.Has(name) {
			return "", wherr.Error("profile already exists")
		}

		return name, nil
	}
}

func promtUserAccount(acc *entities.Account) error {
	fmt.Fprintln(os.Stdout, "Enter new configuration for the Account:")

	for _, input := range []struct {
		prompt      promptui.Prompt
		destination *string
	}{
		{
			prompt: promptui.Prompt{
				Label:     "Name",
				Default:   acc.AuthorName,
				AllowEdit: true,
				Validate: func(s string) error {
					if strings.TrimSpace(s) == "" {
						return wherr.Error("name can't be empty")
					}

					return nil
				},
			},
			destination: &acc.AuthorName,
		},
		{
			prompt: promptui.Prompt{
				Label:     "Short name (optional)",
				Default:   acc.AuthorShortName,
				AllowEdit: true,
			},
			destination: &acc.AuthorShortName,
		},
		{
			prompt: promptui.Prompt{
				Label:     "Athor URL (optional)",
				Default:   acc.AuthorURL,
				AllowEdit: true,
			},
			destination: &acc.AuthorURL,
		},
	} {
		v, err := input.prompt.Run()
		if err != nil {
			return fmt.Errorf("read value: %w", err)
		}

		*input.destination = strings.TrimSpace(v)
	}

	return nil
}

func show(ctx *cli.Context) error {
	cfg, err := config.ReadConfig(ctx.String("profile"))
	if err != nil {
		return err
	}

	acc := cfg.Account()

	if !cfg.Exists() || !acc.Configured() {
		os.Stdout.WriteString("Account is not configured.")

		return nil
	}

	fmt.Fprintf(os.Stdout, "Account for '%s' profile:\n", cfg.Profile)
	fmt.Fprintf(os.Stdout, "\tName: %s\n", acc.AuthorName)
	fmt.Fprintf(os.Stdout, "\tShort name: %s\n", acc.AuthorShortName)
	fmt.Fprintf(os.Stdout, "\tAuthor URL: %s\n", acc.AuthorURL)
	fmt.Fprintf(os.Stdout, "\tToken: %s\n", utils.MaskString(acc.AccessToken))

	return nil
}

func login(ctx *cli.Context) error {
	cfg, err := config.ReadConfig(ctx.String("profile"))
	if err != nil {
		return err
	}

	acc := cfg.Account()

	if !cfg.Exists() || !acc.Configured() {
		os.Stdout.WriteString("Account is not configured.")

		return nil
	}

	_ = 0

	return nil
}

func showToken(ctx *cli.Context) error {
	cfg, err := config.ReadConfig(ctx.String("profile"))
	if err != nil {
		return err
	}

	acc := cfg.Account()

	if !cfg.Exists() || !acc.Configured() {
		os.Stdout.WriteString("Account is not configured.")

		return nil
	}

	if acc.AccessToken == "" {
		os.Stdout.WriteString("No token available, please log in.")

		return nil
	}

	os.Stdout.WriteString(acc.AccessToken)

	return nil
}
