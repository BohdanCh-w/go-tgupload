package entities

type Node struct {
	Tag      string
	Attrs    map[string]string
	Children []any
}
