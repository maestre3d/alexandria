package domain

// UserAggregate required struct for any aggregation operation
type UserAggregate struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	LastName string `json:"last_name"`
	Email    string `json:"email"`
	Gender   string `json:"gender"`
	Locale   string `json:"locale"`
	Role     string `json:"role"`
	Active   string `json:"active"`
}
