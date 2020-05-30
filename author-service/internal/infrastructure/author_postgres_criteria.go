package infrastructure

func QueryCriteriaSQL(query string) string {
	if query == "" {
		return ""
	}
	return `(LOWER(FIRST_NAME) LIKE LOWER('%` + query + `%') OR LOWER(LAST_NAME) LIKE LOWER('%` + query + `%') 
	OR LOWER(DISPLAY_NAME) LIKE LOWER('%` + query + `%'))`
}

func DisplayNameCriteriaSQL(displayName string) string {
	if displayName == "" {
		return ""
	}

	return `LOWER(DISPLAY_NAME) = LOWER('` + displayName + `')`
}

func OwnerIDCriteriaSQL(ownerID string) string {
	if ownerID == "" {
		return ""
	}

	return `OWNER_ID == '` + ownerID + `'`
}

func AndCriteriaSQL(statement string) string {
	return statement + " AND "
}

func OtherCriteriaSQL(statement string) string {
	return statement + " OR "
}
