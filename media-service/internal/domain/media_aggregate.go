package domain

type MediaAggregate struct {
	Title        string `json:"title"`
	DisplayName  string `json:"display_name"`
	Description  string `json:"description"`
	LanguageCode string `json:"language_code"`
	PublisherID  string `json:"publisher_id"`
	AuthorID     string `json:"author_id"`
	PublishDate  string `json:"publish_date"`
	MediaType    string `json:"media_type"`
}

type MediaUpdateAggregate struct {
	Root *MediaAggregate
	ID   string `json:"id"`
	URL  string `json:"url"`
}
