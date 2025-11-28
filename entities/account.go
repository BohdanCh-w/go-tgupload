package entities

type Account struct {
	AuthorName      string
	AuthorShortName string
	AuthorURL       string
	AccessToken     string
}

func (a *Account) Configured() bool {
	return a.AuthorName != "" && a.AuthorShortName != "" && a.AuthorURL == ""
}
