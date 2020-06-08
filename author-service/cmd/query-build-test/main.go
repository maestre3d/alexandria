package main

import (
	"github.com/alexandria-oss/core"
	"github.com/maestre3d/alexandria/author-service/internal/infrastructure"
	"log"
	"strings"
)

func main() {
	filterParams := core.FilterParams{
		"query":          "musk",
		"ownership_type": "root",
		"owner_id":       "123",
	}
	params := core.NewPaginationParams("123", "1")

	b := infrastructure.AuthorBuilder{`SELECT * FROM alexa1.author WHERE `}
	// Criteria map filter -> Query Builder
	for filterType, value := range filterParams {
		// Avoid nil values and comparison computation
		if value == "" {
			continue
		}
		switch {
		case filterType == "query":
			b.Query(value).And()
			continue
		case filterType == "display_name":
			b.DisplayName(value).And()
			continue
		case filterType == "ownership_type":
			b.Ownership(value).And()
			continue
		case filterType == "owner_id":
			b.Owner(value).And()
			continue
		}
	}

	isActive := "TRUE"
	if strings.ToUpper(filterParams["show_disabled"]) == "TRUE" {
		isActive = "FALSE"
	}

	order := ""
	if strings.ToUpper(filterParams["order"]) == "DESC" {
		order = "DESC"
	} else if strings.ToUpper(filterParams["order"]) == "ASC" {
		order = "ASC"
	}

	// Keyset pagination, filtering type binding
	if filterParams["timestamp"] == "false" {
		if params.Token != "" {
			b.Filter("id", ">=", params.Token, isActive).And()
		}

		b.Active(isActive).OrderBy("id", "asc", order)
	} else {
		if params.Token != "" {
			// Timestamp/Most recent by default
			b.Filter("update_time", "<=", params.Token, isActive).And()
		}

		b.Active(isActive).OrderBy("update_time", "desc", order)
	}
	b.Limit(params.Size)
	log.Print(b.Statement)
}
