package infrastructure

import "fmt"

func QueryCriteriaSQL(query string) string {
	if query == "" {
		return ""
	}
	return `(LOWER(first_name) LIKE LOWER('%` + query + `%') OR LOWER(last_name) LIKE LOWER('%` + query + `%') 
	OR LOWER(display_name) LIKE LOWER('%` + query + `%'))`
}

func DisplayNameCriteriaSQL(displayName string) string {
	if displayName == "" {
		return ""
	}

	return `LOWER(DISPLAY_NAME) = LOWER('` + displayName + `')`
}

func OwnershipCriteriaSQL(ownershipType string) string {
	if ownershipType == "" {
		return ""
	}

	return `ownership_type = LOWER('` + ownershipType + `)'`
}

func OwnerCriteriaSQL(ownerID string) string {
	if ownerID == "" {
		return ""
	}

	return fmt.Sprintf(`external_id IN (SELECT fk_author FROM alexa1.author_user WHERE "user" = '%s')`, ownerID)
}

func AndCriteriaSQL(statement string) string {
	return statement + " AND "
}

func OtherCriteriaSQL(statement string) string {
	return statement + " OR "
}
