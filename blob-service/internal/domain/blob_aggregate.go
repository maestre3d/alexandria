package domain

type BlobAggregate struct {
	RootID    string `json:"root_id"`
	Service   string `json:"service"`
	BlobType  string `json:"blob_type"`
	Extension string `json:"extension"`
	Size      int64  `json:"size"`
	Content   File   `json:"content"`
}
