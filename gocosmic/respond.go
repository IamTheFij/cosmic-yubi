package gocosmic

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"git.iamthefij.com/iamthefij/slog"
)

// WriteResponseToStdout will write a PluginResponse from the plugin to the launcher over Stdout
func WriteResponseToStdout(response PluginResponse) error {
	multiWriter := io.MultiWriter(os.Stdout, slog.LoggerDebug.Writer())
	return WriteResponse(multiWriter, response)
}

// WriteResponse will write a PluginResponse from the plugin to the launcher to a provided writer
func WriteResponse(writer io.Writer, response PluginResponse) error {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed marshalling response as JSON: %w", err)
	}

	writer.Write(jsonResponse)
	fmt.Fprintf(writer, "\r\n")

	return nil
}
