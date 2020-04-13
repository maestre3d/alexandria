package infrastructure

import (
	"fmt"
	"strings"
)

func QueryCriteria(query string) string {
	if query == "" {
		return ""
	}
	return `(LOWER(TITLE) LIKE LOWER('%` + query + `%') OR LOWER(DISPLAY_NAME) LIKE LOWER('%` + query + `%') OR 
	LOWER(MEDIA_TYPE::TEXT) LIKE LOWER('%` + query + `%'))`
}

func MediaTypeCriteria(mediaType string) string {
	// Required enum validation
	if mediaType == "" {
		return ""
	}
	return fmt.Sprintf(`MEDIA_TYPE = '%s'`, strings.ToUpper(mediaType))
}

func AuthorCriteria(authorID string) string {
	// Required UUID validation
	if authorID == "" {
		return ""
	}
	return fmt.Sprintf(`AUTHOR_ID = '%s'`, authorID)
}

func PublisherCriteria(userID string) string {
	// Required UUID validation
	if userID == "" {
		return ""
	}
	return fmt.Sprintf(`USER_ID = '%s'`, userID)
}

func AndCriteria(statement string) string {
	return statement + " AND "
}

func OtherCriteria(statement string) string {
	return statement + " OR "
}
