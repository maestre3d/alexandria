package domain

// Relation-table(s) by de-normalization w/ custom indexes

// Categories ordered by root ID
// Cassandra e.g. map<root_id | timestamp, map<category_id, name>>
type CategoryByRoot struct {
	RootID       string `json:"root_id"`
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	CreateTime   string `json:"create_time"`
}
