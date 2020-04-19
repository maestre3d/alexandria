package infrastructure

import (
	"fmt"
	"strings"
)

func QueryCriteriaSQL(query string) string {
	if query == "" {
		return ""
	}
	return `(LOWER(TITLE) LIKE LOWER('%` + query + `%') OR LOWER(DISPLAY_NAME) LIKE LOWER('%` + query + `%') OR 
	LOWER(MEDIA_TYPE::TEXT) LIKE LOWER('%` + query + `%'))`
}

func MediaTypeCriteriaSQL(mediaType string) string {
	// Required enum validation
	if mediaType == "" {
		return ""
	}
	return fmt.Sprintf(`MEDIA_TYPE = '%s'`, strings.ToUpper(mediaType))
}

func AuthorCriteriaSQL(authorID string) string {
	// Required UUID validation
	if authorID == "" {
		return ""
	}
	return fmt.Sprintf(`AUTHOR_ID = '%s'`, authorID)
}

func PublisherCriteriaSQL(userID string) string {
	// Required UUID validation
	if userID == "" {
		return ""
	}
	return fmt.Sprintf(`USER_ID = '%s'`, userID)
}

func TitleCriteriaSQL(title string) string {
	if title == "" {
		return ""
	}

	return `LOWER(TITLE) = LOWER('` + title + `')`
}

func AndCriteriaSQL(statement string) string {
	return statement + " AND "
}

func OtherCriteriaSQL(statement string) string {
	return statement + " OR "
}
