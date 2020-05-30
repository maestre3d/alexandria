package domain

type AuthorAggregate struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	OwnerID     string `json:"owner_id"`
	BirthDate   string `json:"birth_date"`
}
