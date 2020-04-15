package infrastructure

func QueryCriteriaSQL(query string) string {
	if query == "" {
		return ""
	}
	return `(LOWER(FIRST_NAME) LIKE LOWER('%` + query + `%') OR LOWER(LAST_NAME) LIKE LOWER('%` + query + `%') 
	OR LOWER(DISPLAY_NAME) LIKE LOWER('%` + query + `%'))`
}

func AndCriteriaSQL(statement string) string {
	return statement + " AND "
}

func OtherCriteriaSQL(statement string) string {
	return statement + " OR "
}
