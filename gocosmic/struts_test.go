package gocosmic_test

import (
	"slices"
	"testing"

	"git.iamthefij.com/iamthefij/cosmic-yubi/gocosmic"
)

func TestSearchResults(t *testing.T) {
	cases := []struct {
		name     string
		allNames []string
		query    string
		expected []string
	}{
		{
			name: "No matches",
			allNames: []string{
				"chicken",
			},
			query:    "cow",
			expected: []string{},
		},
		{
			name: "Exact match preferred over partial",
			allNames: []string{
				"browncow",
				"cow",
			},
			query: "cow",
			expected: []string{
				"cow",
				"browncow",
			},
		},
		{
			name: "Prefix preferred sort",
			allNames: []string{
				"browncow",
				"bluecow",
				"cowabunga",
				"cow",
			},
			query: "cow",
			expected: []string{
				"cow",
				"cowabunga",
				"bluecow",
				"browncow",
			},
		},
		{
			name: "Case insensitive",
			allNames: []string{
				"Cow",
			},
			query: "cow",
			expected: []string{
				"Cow",
			},
		},
		{
			name: "Case sensitive",
			allNames: []string{
				"cow",
			},
			query:    "Cow",
			expected: []string{},
		},
	}

	for _, c := range cases {
		testCase := c

		t.Run(testCase.name, func(t *testing.T) {
			searchResults := gocosmic.PluginSearchResults{}
			// Add possible results
			for i, name := range testCase.allNames {
				searchResults.Add(gocosmic.PluginSearchResult{
					Id:   uint32(i),
					Name: name,
				})
			}

			queryResults := []string{}
			for _, result := range searchResults.Search(testCase.query) {
				queryResults = append(queryResults, result.Name)
			}

			if !slices.Equal(testCase.expected, queryResults) {
				t.Errorf("expected %v, actual %v", testCase.expected, queryResults)
			}
		})
	}

}
