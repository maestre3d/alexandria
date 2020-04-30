package domain

type IAuthorEventBus interface {
	AuthorCreated(author *AuthorEntity) error
	AuthorDeleted(id string) error
}
