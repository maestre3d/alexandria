package domain

type IAuthorEventBus interface {
	AuthorCreated(author *AuthorEntity) error
	AuthorDeleted(ID string) error
}
