package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cboone/fm/internal/types"
)

func TestJSONFormatter_Format(t *testing.T) {
	f := &JSONFormatter{}
	var buf bytes.Buffer

	data := types.MailboxInfo{
		ID:           "mb1",
		Name:         "Inbox",
		Role:         "inbox",
		TotalEmails:  100,
		UnreadEmails: 5,
	}

	if err := f.Format(&buf, data); err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}

	if result["name"] != "Inbox" {
		t.Errorf("expected name=Inbox, got %v", result["name"])
	}
	if result["role"] != "inbox" {
		t.Errorf("expected role=inbox, got %v", result["role"])
	}
}

func TestJSONFormatter_FormatError(t *testing.T) {
	f := &JSONFormatter{}
	var buf bytes.Buffer

	if err := f.FormatError(&buf, "not_found", "email not found", "check the ID"); err != nil {
		t.Fatal(err)
	}

	var result types.AppError
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}

	if result.Error != "not_found" {
		t.Errorf("expected error=not_found, got %s", result.Error)
	}
	if result.Message != "email not found" {
		t.Errorf("expected message='email not found', got %s", result.Message)
	}
	if result.Hint != "check the ID" {
		t.Errorf("expected hint='check the ID', got %s", result.Hint)
	}
}
