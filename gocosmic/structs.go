package gocosmic

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

type ActivateContext struct {
	Id      uint32 `json:"id"`
	Context uint32 `json:"context"`
}

type Request struct {
	Activate        uint32          `json:",omitzero"`
	ActivateContext ActivateContext `json:",omitzero"`
	Complete        uint32          `json:",omitzero"`
	Context         uint32          `json:",omitzero"`
	Exit            bool            `json:",omitzero"`
	Close           bool            `json:",omitzero"`
	Interrupt       bool            `json:",omitzero"`
	Quit            uint32          `json:",omitzero"`
	Search          string          `json:",omitzero"`
}

func (r *Request) UnmarshalJSON(data []byte) error {
	// If we got valid JSON, but just nothing, we return no error but leave an empty struct
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	// Try to parse bare string results
	var bareString string
	if err := json.Unmarshal(data, &bareString); err == nil {
		// Result was a bare string
		switch bareString {
		case "Exit":
			r.Exit = true
		case "Close":
			r.Close = true
		case "Interrupt":
			r.Interrupt = true
		default:
			return fmt.Errorf("unknown result: %s", bareString)
		}

		return nil
	}

	type tempRequest struct {
		Activate        uint32          `json:",omitzero"`
		ActivateContext ActivateContext `json:",omitzero"`
		Complete        uint32          `json:",omitzero"`
		Context         uint32          `json:",omitzero"`
		Quit            uint32          `json:",omitzero"`
		Search          string          `json:",omitzero"`
	}

	temp := tempRequest{}

	// Try to parse into temp struct
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed parsing response: %w", err)
	}

	// Copy values over
	r.Activate = temp.Activate
	r.ActivateContext = temp.ActivateContext
	r.Complete = temp.Complete
	r.Context = temp.Context
	r.Quit = temp.Quit
	r.Search = temp.Search

	return nil
}

type ContextOption struct {
	Id   uint32 `json:"id"`
	Name string `json:"name"`
}

type GpuPreference uint32

// Rust code is an enum containing Default, NonDefault, and SpecificIdx(u32), so maybe a string is better?

type DesktopEntry struct {
	Path          string        `json:"path"`
	GpuPreference GpuPreference `json:"gpu_preference"`
	ActionName    string        `json:"action_name"`
}

type IconSource struct {
	Name string `json:",omitzero"`
	Mime string `json:",omitzero"`
}

type PluginSearchResult struct {
	Id           uint32     `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Keywords     []string   `json:"keywords,omitempty"`
	Icon         IconSource `json:"icon,omitzero"`
	CategoryIcon IconSource `json:"category_icon,omitzero"`
	Window       []uint32   `json:"window,omitempty"`
}

type ResponseContext struct {
	Id      uint32          `json:"id"`
	Options []ContextOption `json:"options"`
}

type PluginResponse struct {
	Append       PluginSearchResult `json:",omitzero"`
	Clear        bool               `json:",omitzero"`
	Close        bool               `json:",omitzero"`
	Context      ResponseContext    `json:",omitzero"`
	Deactivate   bool               `json:",omitzero"`
	DesktopEntry DesktopEntry       `json:",omitzero"`
	Fill         string             `json:",omitzero"`
	Finished     bool               `json:",omitzero"`
}

func (r PluginResponse) MarshalJSON() ([]byte, error) {
	// Handle single string responses
	if r.Clear {
		return []byte(`"Clear"`), nil
	} else if r.Close {
		return []byte(`"Close"`), nil
	} else if r.Deactivate {
		return []byte(`"Deactivate"`), nil
	} else if r.Finished {
		return []byte(`"Finished"`), nil
	}

	type tempResponse struct {
		Append       PluginSearchResult `json:",omitzero"`
		Context      ResponseContext    `json:",omitzero"`
		DesktopEntry DesktopEntry       `json:",omitzero"`
		Fill         string             `json:",omitzero"`
	}

	temp := tempResponse{
		Append:       r.Append,
		Context:      r.Context,
		DesktopEntry: r.DesktopEntry,
		Fill:         r.Fill,
	}

	result, err := json.Marshal(temp)
	if err != nil {
		return result, fmt.Errorf("failed marshalling plugin response: %w", err)
	}

	return result, nil
}

// PluginSearchResults is a struct used to filter and sort search results
// One would create an instance and then add all results to it, then return the sorted list
type PluginSearchResults struct {
	results []PluginSearchResult
}

func (sr PluginSearchResults) IsEmpty() bool {
	return len(sr.results) == 0
}

func (sr *PluginSearchResults) ClearResults() {
	sr.results = []PluginSearchResult{}
}

// Add adds a search result to the list
func (sr *PluginSearchResults) Add(result PluginSearchResult) {
	sr.results = append(sr.results, result)
}

// Get a result by its Id, if not found an empty result is returned
func (sr PluginSearchResults) Get(id uint32) (PluginSearchResult, bool) {
	for _, result := range sr.results {
		if id == result.Id {
			return result, true
		}
	}

	return PluginSearchResult{}, false
}

// Search will return a slice of filtered and sorted search results
func (sr PluginSearchResults) Search(query string) []PluginSearchResult {
	caseSensitive := true
	// If our query is all lower case, we're going to treat our search as case insensitive
	if strings.ToLower(query) == query {
		caseSensitive = false
	}

	// Filter search results for partial match
	results := []PluginSearchResult{}
	for _, result := range sr.results {
		if strings.Contains(result.Name, query) {
			// Append case sensitive match
			results = append(results, result)
		} else if !caseSensitive && strings.Contains(strings.ToLower(result.Name), query) {
			// Append case insensitive match
			results = append(results, result)
		}
	}

	slices.SortFunc(results, func(a, b PluginSearchResult) int {
		aName := a.Name
		bName := b.Name

		if !caseSensitive {
			aName = strings.ToLower(aName)
			bName = strings.ToLower(bName)
		}

		// Prefer exact match
		if aName == query && bName != query {
			return -1
		} else if aName != query && bName == query {
			return 1
		}

		// Look up position of query in name
		aIndex := strings.Index(aName, query)
		bIndex := strings.Index(bName, query)

		// If one of the names contains the string
		if aIndex > 0 || bIndex > 0 {
			if aIndex < 0 {
				// If 'a' doesn't contain the string, then 'b' is the better match
				return 1
			} else if bIndex < 0 {
				// If 'b' doesn't contain the string, then 'a' is the better match
				return -1
			}

			// Sort based on how close to a prefix the match is
			if aIndex < bIndex {
				return -1
			} else if aIndex > bIndex {
				return 1
			}

			// If indicies are the same, we fall through to a string sort
		}

		return strings.Compare(a.Name, b.Name)
	})

	return results
}
