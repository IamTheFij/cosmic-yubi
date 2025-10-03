package gocosmic

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"git.iamthefij.com/iamthefij/slog"
)

var decoder *json.Decoder

// ReadRequestFromStdin will return Requests from the launcher frontend from Stdin
func ReadRequestFromStdin() (Request, error) {
	return ReadRequest(os.Stdin)
}

// ReadRequest will return Requests from the launcher frontend from a provided reader
func ReadRequest(reader io.Reader) (Request, error) {
	var req Request

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

// ReadRequests will read from the reader and send results out on the results channel
func ReadRequests(reader io.Reader, resultsChan chan Request, stopChan chan bool) {
	decoder := json.NewDecoder(reader)
	for {
		select {
		case <-stopChan:
			close(resultsChan)
			return

		default:
			var req Request

			err := decoder.Decode(&req)
			if err != nil {
				if err == io.EOF {
					// If we've reached EOF, then we close the results channel
					close(resultsChan)
					return
				}

				slog.Errorf("error decoding request json %v", err)
				continue
			}

			resultsChan <- req
		}
	}
}
