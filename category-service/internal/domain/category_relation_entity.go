package domain

// Relation-table(s) by de-normalization w/ custom indexes
// Cassandra e.g. map<category_id | timestamp, map<root_id, null>>

type EntityByCategory struct {
	ID   string `json:"category_id"`
	Root string `json:"root_id"`
}

// Cassandra e.g. map<root_id | timestamp, map<category_id, name>>

type CategoryByEntity struct {
	ID         string `json:"root_id"`
	CategoryID string `json:"category_id"`
	Name       string `json:"category_name"`
}
