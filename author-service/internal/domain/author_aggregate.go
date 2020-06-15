package domain

type AuthorAggregate struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	DisplayName   string `json:"display_name"`
	OwnerID       string `json:"owner_id"`
	OwnershipType string `json:"ownership_type"`
}

type AuthorUpdateAggregate struct {
	ID            string
	RootAggregate *AuthorAggregate
	Verified      string
	Picture       string
}
