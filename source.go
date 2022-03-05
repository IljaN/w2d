package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// Source abstracts  away  underlying storage details. It  can transparently construct io.Reader's  for STDIN or
// a remote url. The UnmarshallText interfaces with the arg package.
type Source struct {
	StdIn      bool
	IsUrl      bool
	url        *url.URL
	rawValue   string
	emptyValue string
}

const emptyValue = "-"

// UnmarshalText configures the factory method Reader so that it returns an STDIN reader if available. If no STDIN is
// connected the argument "text" is parsed as URL and a body-reader will be constructed by Reader
//
// This method initializes the struct.
//
func (a *Source) UnmarshalText(text []byte) error {
	a.rawValue = string(text)

	// Enable STDIN if available, ignore everything else in this case
	if a.stdInConnected() {
		a.StdIn = true
		return nil
	}

	// Don't try to parse a URL if text still has the emptyValue
	if a.IsUnset() {
		return nil
	}

	u, err := url.ParseRequestURI(a.rawValue)
	if err != nil {
		return err
	}

	a.IsUrl = true
	a.url = u

	return nil
}

func (a *Source) Reader() (io.Reader, error) {
	if a.StdIn {
		return os.Stdin, nil
	}

	if a.IsUrl {
		resp, err := http.Get(a.url.String())
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	}

	return nil, fmt.Errorf("invalid state")
}

func (a *Source) stdInConnected() bool {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return true
	} else {
		return false
	}
}

func (a *Source) IsUnset() bool {
	return a.rawValue == emptyValue && !a.StdIn && !a.IsUrl
}
