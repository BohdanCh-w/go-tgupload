package entities

type Page struct {
	Title       string
	AuthorName  *string
	AuthorURL   *string
	Description string
	Content     []Node
}
