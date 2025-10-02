package gocosmic

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ReadRequestFromStdin will return Requests from the launcher frontend from Stdin
func ReadRequestFromStdin() (Request, error) {
	return ReadRequest(os.Stdin)
}

// ReadRequest will return Requests from the launcher frontend from a provided reader
func ReadRequest(reader io.Reader) (Request, error) {
	var req Request

	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&req)
	if err != nil {
		if err == io.EOF {
			// If we've reached EOF, then we return it without wrapping
			return req, io.EOF
		}

		return req, fmt.Errorf("error decoding request JSON: %w", err)
	}

	return req, nil
}
