package account

import (
	"fmt"
	"os"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/manifoldco/promptui"
	"github.com/pkg/browser"
	"github.com/urfave/cli/v2"

	"github.com/bohdanch-w/go-tgupload/config"
	"github.com/bohdanch-w/go-tgupload/entities"
	"github.com/bohdanch-w/go-tgupload/integrations/telegraph"
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
				Name:   "setup",
				Action: setup,
			},
			{
				Name:   "login",
				Action: login,
			},
			{
				Name: "web-login",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "text-only",
					},
				},
				Action: webLogin,
			},
			{
				Name:   "token",
				Action: showToken,
			},
			{
				Name:   "validate",
				Action: validate,
			},
			{
				Name:   "sync",
				Action: sync,
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
	if err := promptUserAccount(&acc); err != nil {
		return err
	}

	confirmed := prompter.YN(fmt.Sprintf(
		`Confirm changes to %q profile
 - Name:       %s
 - Short name: %s
 - Author URL: %s
 `,
		cfg.Profile,
		acc.AuthorName,
		acc.AuthorShortName,
		acc.AuthorURL,
	), true)

	if !confirmed {
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

	profiles = append(profiles, "!new")

	selectedProfile := prompter.Choose("Select profile", profiles, profiles[0])
	if selectedProfile != "!new" {
		return selectedProfile, nil
	}

	pfs := hashset.New(profiles...)

	for {
		name := prompter.Prompt("Profile ID", "")
		if name == "" {
			continue
		}

		if pfs.Has(name) {
			return "", wherr.Error("profile already exists")
		}

		return name, nil
	}
}

func promptUserAccount(acc *entities.Account) error {
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
				Label:     "Short name",
				Default:   acc.AuthorShortName,
				AllowEdit: true,
				Validate: func(s string) error {
					if strings.TrimSpace(s) == "" {
						return wherr.Error("name can't be empty")
					}

					return nil
				},
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

	token, err := telegraph.Login(ctx.Context, acc)
	if err != nil {
		return fmt.Errorf("failed to log in: %w", err)
	}

	acc.AccessToken = token

	cfg.SetAccount(acc)

	if err := config.StoreConfig(cfg); err != nil {
		return fmt.Errorf("store config: %w", err)
	}

	return nil
}

func sync(ctx *cli.Context) error {
	cfg, err := config.ReadConfig(ctx.String("profile"))
	if err != nil {
		return err
	}

	acc := cfg.Account()

	if !cfg.Exists() {
		return wherr.Error("Config not exists")
	}

	if acc.AccessToken == "" {
		return wherr.Error("Token is not configured")
	}

	tg, err := telegraph.New(acc)
	if err != nil {
		return fmt.Errorf("init telegraph API: %w", err)
	}

	accInfo, err := tg.Account(
		ctx.Context,
		"author_name",
		"short_name",
		"author_url",
		"page_count",
	)
	if err != nil {
		return fmt.Errorf("get info: %w", err)
	}

	acc.AuthorName = accInfo.AuthorName
	acc.AuthorShortName = accInfo.AuthorShortName
	acc.AuthorURL = accInfo.AuthorURL

	cfg.SetAccount(acc)

	if err := config.StoreConfig(cfg); err != nil {
		return fmt.Errorf("store config: %w", err)
	}

	fmt.Fprintln(os.Stdout, "Account is successfully synced.")

	return nil
}

func webLogin(ctx *cli.Context) error {
	cfg, err := config.ReadConfig(ctx.String("profile"))
	if err != nil {
		return err
	}

	acc := cfg.Account()

	if !cfg.Exists() {
		return wherr.Error("Account is not configured")
	}

	if acc.AccessToken == "" {
		return wherr.Error("Token is not configured")
	}

	tg, err := telegraph.New(acc)
	if err != nil {
		return fmt.Errorf("init telegraph API: %w", err)
	}

	accInfo, err := tg.Account(
		ctx.Context,
		"auth_url",
	)
	if err != nil {
		return fmt.Errorf("get info: %w", err)
	}

	fmt.Fprintln(os.Stdout, "Your login link:", accInfo.AuthURL)

	if !ctx.Bool("text-only") {
		browser.OpenURL(accInfo.AuthURL)
	}

	return nil
}

func validate(ctx *cli.Context) error {
	cfg, err := config.ReadConfig(ctx.String("profile"))
	if err != nil {
		return err
	}

	acc := cfg.Account()

	if !cfg.Exists() {
		return wherr.Error("Account is not configured")
	}

	if acc.AccessToken == "" {
		return wherr.Error("Token is not configured")
	}

	tg, err := telegraph.New(acc)
	if err != nil {
		return fmt.Errorf("init telegraph API: %w", err)
	}

	accInfo, err := tg.Account(
		ctx.Context,
		"author_name",
		"short_name",
		"author_url",
		"page_count",
	)
	if err != nil {
		return fmt.Errorf("get info: %w", err)
	}

	fmt.Fprintln(os.Stdout, "Account is successfully configured:")
	fmt.Fprintf(os.Stdout, "\tName: %s\n", accInfo.AuthorName)
	fmt.Fprintf(os.Stdout, "\tShort name: %s\n", accInfo.AuthorShortName)
	fmt.Fprintf(os.Stdout, "\tAuthor URL: %s\n", accInfo.AuthorURL)
	fmt.Fprintf(os.Stdout, "\tArticles published: %d\n", accInfo.PageCount)

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
