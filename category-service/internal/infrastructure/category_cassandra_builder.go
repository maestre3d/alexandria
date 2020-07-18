package infrastructure

import (
	"fmt"
	"strings"
)

type CategoryCassandraBuilder struct {
	Statement string
}

func (b *CategoryCassandraBuilder) Name(query string) *CategoryCassandraBuilder {
	b.Statement += fmt.Sprintf(`category_name = '%s'`, strings.Title(query))
	return b
}

func (b *CategoryCassandraBuilder) Active(flag bool) *CategoryCassandraBuilder {
	if flag {
		b.Statement += `active = True`
		return b
	}

	b.Statement += `active = False`
	return b
}

func (b *CategoryCassandraBuilder) Limit(size int) *CategoryCassandraBuilder {
	b.Statement += fmt.Sprintf(" LIMIT %d", size)
	return b
}

func (b *CategoryCassandraBuilder) And() *CategoryCassandraBuilder {
	b.Statement += ` AND `
	return b
}
