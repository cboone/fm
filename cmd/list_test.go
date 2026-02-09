package cmd

import "testing"

func TestParseSort_Default(t *testing.T) {
	field, asc, err := parseSort("receivedAt desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if field != "receivedAt" {
		t.Errorf("expected field=receivedAt, got %s", field)
	}
	if asc {
		t.Error("expected ascending=false for desc")
	}
}

func TestParseSort_Ascending(t *testing.T) {
	field, asc, err := parseSort("sentAt asc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if field != "sentAt" {
		t.Errorf("expected field=sentAt, got %s", field)
	}
	if !asc {
		t.Error("expected ascending=true for asc")
	}
}

func TestParseSort_AscendingCaseInsensitive(t *testing.T) {
	field, asc, err := parseSort("subject ASC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if field != "subject" {
		t.Errorf("expected field=subject, got %s", field)
	}
	if !asc {
		t.Error("expected ascending=true for ASC")
	}
}

func TestParseSort_FieldOnly(t *testing.T) {
	field, asc, err := parseSort("from")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if field != "from" {
		t.Errorf("expected field=from, got %s", field)
	}
	if asc {
		t.Error("expected ascending=false when direction omitted")
	}
}

func TestParseSort_Empty(t *testing.T) {
	field, asc, err := parseSort("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if field != "receivedAt" {
		t.Errorf("expected default field=receivedAt, got %s", field)
	}
	if asc {
		t.Error("expected ascending=false for empty input")
	}
}

func TestParseSort_ExtraWhitespace(t *testing.T) {
	field, asc, err := parseSort("  receivedAt   asc  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if field != "receivedAt" {
		t.Errorf("expected field=receivedAt, got %s", field)
	}
	if !asc {
		t.Error("expected ascending=true")
	}
}

func TestParseSort_UnknownDirection(t *testing.T) {
	_, _, err := parseSort("receivedAt up")
	if err == nil {
		t.Error("expected error for unknown sort direction")
	}
}

func TestParseSort_InvalidField(t *testing.T) {
	_, _, err := parseSort("invalid desc")
	if err == nil {
		t.Error("expected error for invalid sort field")
	}
}

func TestParseSort_FieldCaseNormalization(t *testing.T) {
	field, _, err := parseSort("ReceivedAt desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if field != "receivedAt" {
		t.Errorf("expected field=receivedAt after normalization, got %s", field)
	}
}

func TestParseSort_ColonSyntax(t *testing.T) {
	field, asc, err := parseSort("receivedAt:asc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if field != "receivedAt" {
		t.Errorf("expected field=receivedAt, got %s", field)
	}
	if !asc {
		t.Error("expected ascending=true for :asc")
	}
}

func TestParseSort_ColonSyntaxDesc(t *testing.T) {
	field, asc, err := parseSort("sentAt:desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if field != "sentAt" {
		t.Errorf("expected field=sentAt, got %s", field)
	}
	if asc {
		t.Error("expected ascending=false for :desc")
	}
}
