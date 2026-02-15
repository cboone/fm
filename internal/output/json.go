package output

import (
	"encoding/json"
	"io"

	"github.com/cboone/fm/internal/types"
)

// JSONFormatter outputs data as indented JSON.
type JSONFormatter struct{}

func (f *JSONFormatter) Format(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func (f *JSONFormatter) FormatError(w io.Writer, code string, message string, hint string) error {
	return f.Format(w, types.AppError{
		Error:   code,
		Message: message,
		Hint:    hint,
	})
}
