package domain

import (
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-playground/validator/v10"
	"io"
	"math"
	"strings"
	"time"
)

const (
	Media  = "media"
	Author = "author"
	User   = "user"
)

type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

type Blob struct {
	ID         string `json:"id" docstore:"id" validate:"required"`
	Service    string `json:"service" docstore:"service" validate:"required,oneof=media author user"`
	Name       string `json:"name" docstore:"name" validate:"required,min=1,max=512"`
	Size       int64  `json:"size" docstore:"size"`
	BlobType   string `json:"blob_type" docstore:"blob_type" validate:"required"`
	Extension  string `json:"extension" docstore:"extension" validate:"required,min=1,max=8"`
	Url        string `json:"url" docstore:"url" validate:"max=2048"`
	CreateTime int64  `json:"create_time" docstore:"create_time"`
	UpdateTime int64  `json:"update_time" docstore:"update_time"`
	Content    File   `json:"-" docstore:"-"`
}

func NewBlob(rootID, service, blobType, extension string, size int64) *Blob {
	blob := &Blob{
		ID:         rootID,
		Service:    strings.ToLower(service),
		Name:       rootID + "." + extension,
		Size:       size,
		BlobType:   blobType,
		Extension:  strings.ToLower(extension),
		Url:        "",
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}

	blob.Url = fmt.Sprintf("https://%s/%s/%s/%s", StorageDomain, StoragePath, blob.Service, blob.Name)
	// Attach Service_ID as first 4 bytes in Root_ID to avoid entity collision in logging tables
	blob.ID = GetServiceID(blob.Service) + blob.ID

	return blob
}

func GetServiceID(service string) string {
	service = strings.ToLower(service)
	switch service {
	case User:
		return "0001"
	case Author:
		return "0002"
	case Media:
		return "0003"
	default:
		return "0000"
	}
}

func (e Blob) IsValid() error {
	// Struct validation
	validate := validator.New()

	err := validate.Struct(e)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch {
			case err.Tag() == "required":
				return exception.NewErrorDescription(exception.RequiredField,
					fmt.Sprintf(exception.RequiredFieldString, strings.ToLower(err.Field())))
			case err.Tag() == "alphanum" || err.Tag() == "alpha" || err.Tag() == "url":
				return exception.NewErrorDescription(exception.InvalidFieldFormat,
					fmt.Sprintf(exception.InvalidFieldFormatString, strings.ToLower(err.Field()), err.Tag()))
			case err.Tag() == "max" || err.Tag() == "min":
				field := strings.ToLower(err.Field())
				min := "1"
				max := "n"
				switch field {
				case "name":
					max = "512"
				case "extension":
					max = "8"
				case "url":
					max = "2048"
				}

				return exception.NewErrorDescription(exception.InvalidFieldRange,
					fmt.Sprintf(exception.InvalidFieldRangeString, field, min, max))
			case err.Tag() == "oneof":
				return exception.NewErrorDescription(exception.InvalidFieldFormat,
					fmt.Sprintf(exception.InvalidFieldFormatString, strings.ToLower(err.Field()),
						"["+err.Param()+")"))
			}
		}

		return err
	}

	// go's file.size is in bytes
	var maxSize int64
	// Default Max 10MB
	maxSize = 10 * (int64(math.Pow(1024, 2)))

	switch e.Service {
	case User:
		if e.Size > maxSize {
			return exception.NewErrorDescription(exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString,
				"file", "1 B", "10 MB"))
		}
		if e.Extension != "jpeg" {
			return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
				"file", "jpeg"))
		}
	case Author:
		if e.Size > maxSize {
			return exception.NewErrorDescription(exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString,
				"file", "1 B", "10 MB"))
		}
		if e.Extension != "jpeg" {
			return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
				"file", "jpeg"))
		}
	case Media:
		// Using IANA MIME types
		switch e.BlobType {
		case "application":
			// Max 25MB
			maxSize = 25 * (int64(math.Pow(1024, 2)))
			if e.Size > maxSize {
				return exception.NewErrorDescription(exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString,
					"file", "1 B", "25 MB"))
			}
			if e.Extension != "pdf" {
				return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
					"file", "pdf"))
			}
		case "audio":
			// Max 256MB
			maxSize = 256 * (int64(math.Pow(1024, 2)))
			if e.Size > maxSize {
				return exception.NewErrorDescription(exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString,
					"file", "1 B", "10 MB"))
			}
			if e.Extension != "mpeg" && e.Extension != "vorbis" && e.Extension != "aac" && e.Extension != "mp4" && e.Extension != "ogg" {
				return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
					"file", fmt.Sprintf("[%s, %s, %s, %s, %s)", "mpeg", "vorbis", "aac", "mp4", "ogg")))
			}
		case "video":
			// Max 8GB
			maxSize = 8192 * (int64(math.Pow(1024, 2)))
			if e.Size > maxSize {
				return exception.NewErrorDescription(exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString,
					"file", "1 B", "10 MB"))
			}

			extension := strings.ToLower(e.Extension)
			if extension != "h264" && extension != "mp4" && extension != "mpeg" && extension != "ogg" {
				return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
					"file", fmt.Sprintf("[%s, %s, %s, %s)", "h264", "mp4", "mpeg", "ogg")))
			}
		default:
			return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
				"file", "document, audio or video"))
		}
	}

	return nil
}
