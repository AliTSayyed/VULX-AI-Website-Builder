package domain

/*
a page is a generic type that will be used for returning items in a paginated manner
token will contian the next
*/

type Page[T any] struct {
	Items   []T
	Token   string
	HasMore bool
}
