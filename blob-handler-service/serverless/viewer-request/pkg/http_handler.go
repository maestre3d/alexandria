package pkg

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Dimension struct {
	Height int
	Width  int
}

var (
	allowedDimensions = []Dimension{{
		Height: 100,
		Width:  100,
	}, {
		Height: 200,
		Width:  200,
	}, {
		Height: 300,
		Width:  300,
	}, {
		Height: 400,
		Width:  400,
	}}
	defaultDimension = Dimension{
		Height: 200,
		Width:  200,
	}
	variance      = 20
	webpExtension = "webp"
)

func GetContentHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("d") == "" {
		// If dimension was not given, then just pass the request
		return
	}

	// Dimension will come as "1280x720", split it
	dimensionMatch := strings.Split(r.URL.Query().Get("d"), "x")
	if len(dimensionMatch) < 2 {
		// If dimension is not valid, pass request
		return
	}

	width64, err := strconv.ParseInt(dimensionMatch[0], 10, 32)
	if err != nil {
		return
	}
	width := int(width64)

	height64, err := strconv.ParseInt(dimensionMatch[1], 10, 32)
	if err != nil {
		return
	}
	height := int(height64)

	isAllowed := false

	// Segregate path, key and extension
	// e.g. "/alexandria/media/cover/123.jpeg" -> "/alexandria/media/cover/", "123.jpeg"
	prefix, file := segregateURI(r.URL.Path)

	variancePercent := (variance / 100)
	for _, dimension := range allowedDimensions {
		minWidth := dimension.Width - (dimension.Width * variancePercent)
		maxWidth := dimension.Width - (dimension.Width * variancePercent)
		if width >= minWidth && width >= maxWidth {
			width = dimension.Width
			height = dimension.Height
			isAllowed = true
			break
		}
	}

	// Set default values if not valid
	if !isAllowed {
		width = defaultDimension.Width
		height = defaultDimension.Height
	}

	acceptHeader := r.Header.Get("accept")

	url := make([]string, 0)
	url = append(url, prefix, fmt.Sprintf("%dx%d", width, height))

	if acceptHeader != "" {
		url = append(url, webpExtension)
	}

	url = append(url, file)
	r.URL.Path = strings.Join(url, "/")
	return
}

func segregateURI(uri string) (string, string) {
	match := strings.Split(uri, "/")
	// 1st return value = prefix
	// 2nd return value = file name and extension
	return strings.Join(match[:len(match)-1], "/"), match[len(match)-1]
}
