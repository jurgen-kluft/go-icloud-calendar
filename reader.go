package icalendar

import (
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
	textualContent := string(response.Body)

	//return the file that contains the info
	return textualContent, nil
}

func ReadingFromURL(url string) Reader {
	return &readFromURL{url: url}
}
