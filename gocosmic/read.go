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
func ReadRequests(reader io.Reader, requests chan Request, done chan bool) {
	decoder := json.NewDecoder(reader)
	for {
		select {
		case <-done:
			slog.Debugf("reader signal recieved on done, closing requests")
			close(requests)
			return

		default:
			var req Request

			err := decoder.Decode(&req)
			if err != nil {
				if err == io.EOF {
					// If we've reached EOF, then we close the results channel
					slog.Debugf("reader has reached EOF, closing requests channel and exiting read loop")
					close(requests)
					return
				}

				slog.Errorf("error decoding request json %v", err)
				continue
			}

			// If the request is asking us to close, we can signal done
			if req.Close || req.Exit || req.Interrupt {
				// Send the done signal and close the channel
				slog.Debugf("reader decoded Close/Exit/Interrupt request, closing requests channel")
				close(requests)
			} else {
				requests <- req
				slog.Debugf("request sent from reader to requests channel")
			}
		}
	}
}
