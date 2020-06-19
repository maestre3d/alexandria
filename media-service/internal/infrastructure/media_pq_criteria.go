package infrastructure

import (
	"fmt"
	"strings"
)

// PQ Query builder
type MediaQuery struct {
	Statement string
}

func (b *MediaQuery) Like(query string) *MediaQuery {
	if query == "" {
		return b
	}

	b.Statement += `LOWER(title) LIKE LOWER('%` + query + `%') OR LOWER(display_name) LIKE LOWER('%` + query + `%')`
	return b
}

func (b *MediaQuery) Language(lang string) *MediaQuery {
	if lang == "" {
		return b
	}

	b.Statement += `language_code = '` + lang + `'`
	return b
}

func (b *MediaQuery) Publisher(id string) *MediaQuery {
	if id == "" {
		return b
	}

	b.Statement += `publisher_id = '` + id + `'`
	return b
}

func (b *MediaQuery) Author(id string) *MediaQuery {
	if id == "" {
		return b
	}

	b.Statement += `author_id = '` + id + `'`
	return b
}

func (b *MediaQuery) MediaType(media string) *MediaQuery {
	if media == "" {
		return b
	}

	b.Statement += `media_type = '` + media + `'`
	return b
}

// Filter
// Returns a query to filter useful fields like timestamp, id, or total_views
/*
key = field,
op = operator,
id = entity external_id,
state = is entity active
*/
func (b *MediaQuery) Filter(key, op, id, state string) *MediaQuery {
	state = strings.ToUpper(state)
	b.Statement += fmt.Sprintf(`%s %s (SELECT %s FROM alexa1.media WHERE external_id = '%s' AND active = %s)`,
		key, op, key, id, state)
	return b
}

// Generic SQL

// Active return a query to search by entity's state
func (b *MediaQuery) Active(state string) *MediaQuery {
	b.Statement += "active = " + state
	return b
}

// Raw returns a query with the raw SQL query
func (b *MediaQuery) Raw(statement string) *MediaQuery {
	b.Statement += statement
	return b
}

// OrderBy returns a query for ordering
/*
key = field,
def = default order,
param = sorting from params, will replace default value
*/
func (b *MediaQuery) OrderBy(key, def, param string) *MediaQuery {
	if param != "" {
		b.Statement += fmt.Sprintf(` ORDER BY %s %s`, key, param)
		return b
	}

	b.Statement += fmt.Sprintf(` ORDER BY %s %s`, key, strings.ToUpper(def))
	return b
}

// Limit returns a query with a limiter, useful for pagination
func (b *MediaQuery) Limit(limit int) *MediaQuery {
	b.Statement += fmt.Sprintf(` FETCH FIRST %d ROWS ONLY`, limit)
	return b
}

// And returns a query with the AND statement
func (b *MediaQuery) And() *MediaQuery {
	b.Statement += " AND "
	return b
}

// Or returns a query with the OR statement
func (b *MediaQuery) Or() *MediaQuery {
	b.Statement += " OR "
	return b
}
