package gocosmic_test

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"git.iamthefij.com/iamthefij/cosmic-yubi/gocosmic"
)

func TestReadRequests(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected []gocosmic.Request
	}{
		{
			name:  "Simple search",
			input: `{"Search": "search"}`,
			expected: []gocosmic.Request{
				{Search: "search"},
			},
		},
		{
			name: "Search activate",
			input: `{"Search": "search"}
			{"Activate": 1}`,
			expected: []gocosmic.Request{
				{Search: "search"},
				{Activate: 1},
			},
		},
	}

	for _, c := range cases {
		c := c

		t.Run(c.name, func(t *testing.T) {
			inputStream := strings.NewReader(c.input)
			results := []gocosmic.Request{}
			for {
				request, err := gocosmic.ReadRequest(inputStream)
				if err == io.EOF {
					break
				}

				if err != nil {
					t.Fatalf("Failed parsing input stream")
				}

				results = append(results, request)
			}

			if !reflect.DeepEqual(c.expected, results) {
				t.Fatalf("expected: %+v got: %+v", c.expected, results)
			}
		})
	}
}
