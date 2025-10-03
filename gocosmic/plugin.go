package gocosmic

import (
	"os"

	"git.iamthefij.com/iamthefij/slog"
)

type RequestHandler func(*PluginSearchResults, Request) error

type Plugin struct {
	requests chan Request
	done     chan bool
	errors   chan error
	results  PluginSearchResults
	handler  RequestHandler
}

func NewPlugin(handler RequestHandler) Plugin {
	return Plugin{
		requests: make(chan Request),
		done:     make(chan bool),
		errors:   make(chan error),
		results:  PluginSearchResults{},
		handler:  handler,
	}
}

func (p Plugin) Run() {
	// Start reader from stdin
	go ReadRequests(os.Stdin, p.requests, p.done)

	// Listen for responses
	for {
		select {
		case request, ok := <-p.requests:
			if !ok {
				// If the requests channel is closed, we can exit
				return
			}

			// If the request is asking us to close, we can signal done
			if request.Close || request.Exit || request.Interrupt {
				// Send the done signal and close the channel
				p.done <- true
				return
			}

			err := p.handler(&p.results, request)
			slog.OnErrFatalf(err, "error handling request")
		case err := <-p.errors:
			slog.OnErrFatalf(err, "failed")
		}
	}
}
