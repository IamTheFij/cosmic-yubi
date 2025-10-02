package gocosmic

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// WriteResponseToStdout will write a PluginResponse from the plugin to the launcher over Stdout
func WriteResponseToStdout(response PluginResponse) error {
	return WriteResponse(os.Stdout, response)
}

// WriteResponse will write a PluginResponse from the plugin to the launcher to a provided writer
func WriteResponse(writer io.Writer, response PluginResponse) error {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed marshalling response as JSON: %w", err)
	}

	writer.Write(jsonResponse)

	return nil
}
