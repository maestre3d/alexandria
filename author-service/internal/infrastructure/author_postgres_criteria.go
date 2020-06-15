package infrastructure

import (
	"fmt"
	"strings"
)

type AuthorBuilder struct {
	Statement string
}

// Query returns a query from most important fields
func (b *AuthorBuilder) Query(query string) *AuthorBuilder {
	if query == "" {
		return b
	}

	b.Statement += `(LOWER(first_name) LIKE LOWER('%` + query + `%') OR LOWER(last_name) LIKE LOWER('%` + query + `%') 
	OR LOWER(display_name) LIKE LOWER('%` + query + `%'))`

	return b
}

// DisplayName returns a query to search by display_name field
func (b *AuthorBuilder) DisplayName(displayName string) *AuthorBuilder {
	if displayName == "" {
		return b
	}

	b.Statement += `LOWER(DISPLAY_NAME) = LOWER('` + displayName + `')`
	return b
}

// Ownership returns a query to search by ownership_type field
func (b *AuthorBuilder) Ownership(ownershipType string) *AuthorBuilder {
	if ownershipType == "" {
		return b
	}

	b.Statement += `ownership_type = '` + ownershipType + `'`
	return b
}

// Owner returns a query to search by owner from owner_pool
func (b *AuthorBuilder) Owner(ownerID string) *AuthorBuilder {
	if ownerID == "" {
		return b
	}

	b.Statement += fmt.Sprintf(`owner_id = '%s'`, ownerID)
	return b
}

// Filter returns a query to filter useful fields like timestamp, id, or total_views
/*
key = field,
op = operator,
id = entity external_id,
state = is entity active
*/
func (b *AuthorBuilder) Filter(key, op, id, state string) *AuthorBuilder {
	state = strings.ToUpper(state)
	b.Statement += fmt.Sprintf(`%s %s (SELECT %s FROM alexa1.author WHERE external_id = '%s' AND active = %s)`,
		key, op, key, id, state)
	return b
}

/* Generic SQL */

// Active return a query to search by entity's state
func (b *AuthorBuilder) Active(state string) *AuthorBuilder {
	b.Statement += "active = " + state
	return b
}

// Raw returns a query with the raw SQL query
func (b *AuthorBuilder) Raw(statement string) *AuthorBuilder {
	b.Statement += statement
	return b
}

// OrderBy returns a query for ordering
/*
key = field,
def = default order,
param = sorting from params, will replace default value
*/
func (b *AuthorBuilder) OrderBy(key, def, param string) *AuthorBuilder {
	if param != "" {
		b.Statement += fmt.Sprintf(` ORDER BY %s %s`, key, param)
		return b
	}

	b.Statement += fmt.Sprintf(` ORDER BY %s %s`, key, strings.ToUpper(def))
	return b
}

// Limit returns a query with a limiter, useful for pagination
func (b *AuthorBuilder) Limit(limit int) *AuthorBuilder {
	b.Statement += fmt.Sprintf(` FETCH FIRST %d ROWS ONLY`, limit)
	return b
}

// And returns a query with the AND statement
func (b *AuthorBuilder) And() *AuthorBuilder {
	b.Statement += " AND "
	return b
}

// Or returns a query with the OR statement
func (b *AuthorBuilder) Or() *AuthorBuilder {
	b.Statement += " OR "
	return b
}
