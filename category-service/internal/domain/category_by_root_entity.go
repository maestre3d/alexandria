package domain

// Relation-table(s) by de-normalization w/ sorted map

// Categories of a root entity
//
// Cassandra e.g. map<root_id | timestamp, map<category_id, name>>
type CategoryByRoot struct {
	RootID       string            `json:"root_id"`
	CategoryList map[string]string `json:"category_list"`
}

func NewCategoryByRoot(rootID, categoryID, categoryName string) *CategoryByRoot {
	return &CategoryByRoot{
		RootID:       rootID,
		CategoryList: map[string]string{categoryID: categoryName},
	}
}
