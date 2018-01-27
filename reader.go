package icalendar

import (
	"bytes"
	"net/http"
)

// Reader can be any location to get ICS data (URL, File, ...)
type Reader interface {
	Read() (string, error)
}

type readFromURL struct {
	url string
}

func (r *readFromURL) Read() (string, error) {
	response, err := http.Get(r.url)
	if err != nil {
		return "", err
	}

	// close the response body
	defer response.Body.Close()

	// copy the response from response in a string
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	textualContent := buf.String()
	//return the file that contains the info
	return textualContent, nil
}

// ReadingFromURL returns an instance that can download content from URL
func ReadingFromURL(url string) Reader {
	return &readFromURL{url: url}
}
