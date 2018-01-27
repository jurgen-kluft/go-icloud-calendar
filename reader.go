package icalendar

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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

	if strings.HasPrefix(textualContent, "BEGIN:VCALENDAR") {
		//return the file that contains the info
		return textualContent, nil
	}
	return "", fmt.Errorf("Content has wrong format")
}

// ReadingFromURL returns an instance that can download content from URL
func readingFromURL(url string) Reader {
	return &readFromURL{url: url}
}

type readFromFile struct {
	filepath string
}

func (r *readFromFile) Read() (string, error) {
	calBytes, err := ioutil.ReadFile(r.filepath)
	if err != nil {
		return "", fmt.Errorf("Failed to read calendar file ( %s )", err)
	}
	return string(calBytes), nil
}

// ReadingFromFile returns a Reader instance that will read from file 'filepath'
func readingFromFile(filepath string) Reader {
	return &readFromFile{filepath: filepath}
}
