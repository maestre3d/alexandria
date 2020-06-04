package domain

type AuthorAggregate struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	DisplayName   string `json:"display_name"`
	OwnershipType string `json:"ownership_type"`
	OwnerID       string `json:"owner_id"`
}

type ownersAggregate []*Owner

type AuthorUpdateAggregate struct {
	ID            string
	RootAggregate *AuthorAggregate
	Owners        ownersAggregate
	Verified      string
	Picture       string
}
