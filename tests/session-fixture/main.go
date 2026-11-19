// Session-fixture serves authentication failures to isolated CLI tests.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
)

func main() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"status":401,"detail":"offline session fixture"}`)
	}))
	defer server.Close()
	if _, err := fmt.Fprintln(os.Stdout, server.URL+"/session"); err != nil {
		return
	}
	// Closing the parent's control pipe ends the owned server without PID signals.
	_, _ = io.Copy(io.Discard, os.Stdin)
}
