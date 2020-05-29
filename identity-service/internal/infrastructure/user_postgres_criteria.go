package infrastructure

func UserQueryCriteria(query string) string {
	return `(LOWER(USERNAME) LIKE LOWER('%` + query + `%') OR LOWER(NAME) LIKE LOWER('%` + query + `%') 
	OR LOWER(LAST_NAME) LIKE LOWER('%` + query + `%'))`
}

func AndCriteriaSQL(statement string) string {
	return statement + " AND "
}

func OtherCriteriaSQL(statement string) string {
	return statement + " OR "
}
