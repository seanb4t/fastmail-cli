package fastmail

import (
	"encoding/json"
	"io"
)

// ExportJSONL writes emails in JSON Lines format to the given writer.
// Each email is written as a single JSON object on its own line.
// This format is streaming-friendly and efficient for large exports.
func ExportJSONL(w io.Writer, emails []Email) error {
	enc := json.NewEncoder(w)
	for _, email := range emails {
		if err := enc.Encode(email); err != nil {
			return err
		}
	}
	return nil
}
