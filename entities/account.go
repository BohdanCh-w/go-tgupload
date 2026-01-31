package entities

type Account struct {
	AuthorName      string
	AuthorShortName string
	AuthorURL       string
	AccessToken     string
	AuthURL         string
	PageCount       uint
}

func (a *Account) Configured() bool {
	return a.AuthorName != ""
}
