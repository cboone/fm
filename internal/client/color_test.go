package client

import (
	"testing"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
)

// --- ParseFlagColor tests ---

func TestParseFlagColor_ValidNames(t *testing.T) {
	cases := []struct {
		input string
		want  FlagColor
	}{
		{"red", FlagColorRed},
		{"orange", FlagColorOrange},
		{"yellow", FlagColorYellow},
		{"green", FlagColorGreen},
		{"blue", FlagColorBlue},
		{"purple", FlagColorPurple},
		{"gray", FlagColorGray},
	}
	for _, tc := range cases {
		got, err := ParseFlagColor(tc.input)
		if err != nil {
			t.Errorf("ParseFlagColor(%q) returned error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseFlagColor(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestParseFlagColor_CaseInsensitive(t *testing.T) {
	cases := []string{"Red", "RED", "Orange", "BLUE", "Gray"}
	for _, input := range cases {
		_, err := ParseFlagColor(input)
		if err != nil {
			t.Errorf("ParseFlagColor(%q) should accept case-insensitive input, got: %v", input, err)
		}
	}
}

func TestParseFlagColor_Invalid(t *testing.T) {
	cases := []string{"", "magenta", "123", "noir"}
	for _, input := range cases {
		_, err := ParseFlagColor(input)
		if err == nil {
			t.Errorf("ParseFlagColor(%q) should return error for invalid color", input)
		}
	}
}

// --- FlagColor.String tests ---

func TestFlagColorString(t *testing.T) {
	cases := []struct {
		color FlagColor
		want  string
	}{
		{FlagColorRed, "red"},
		{FlagColorOrange, "orange"},
		{FlagColorYellow, "yellow"},
		{FlagColorGreen, "green"},
		{FlagColorBlue, "blue"},
		{FlagColorPurple, "purple"},
		{FlagColorGray, "gray"},
	}
	for _, tc := range cases {
		got := tc.color.String()
		if got != tc.want {
			t.Errorf("FlagColor(%d).String() = %q, want %q", tc.color, got, tc.want)
		}
	}
}

// --- FlagColor.Patch tests ---

func TestFlagColorPatch_AllColors(t *testing.T) {
	// Verify the 3-bit encoding per IETF spec:
	// Value = Bit0 + 2*Bit1 + 4*Bit2
	cases := []struct {
		color FlagColor
		bit0  bool
		bit1  bool
		bit2  bool
	}{
		{FlagColorRed, false, false, false},   // 0 = 000
		{FlagColorOrange, true, false, false}, // 1 = 001
		{FlagColorYellow, false, true, false}, // 2 = 010
		{FlagColorGreen, true, true, false},   // 3 = 011
		{FlagColorBlue, false, false, true},   // 4 = 100
		{FlagColorPurple, true, false, true},  // 5 = 101
		{FlagColorGray, false, true, true},    // 6 = 110
	}
	for _, tc := range cases {
		patch := tc.color.Patch()
		if len(patch) != 3 {
			t.Fatalf("FlagColor(%d).Patch() has %d keys, want 3", tc.color, len(patch))
		}
		checkBit(t, tc.color, patch, "keywords/$MailFlagBit0", tc.bit0)
		checkBit(t, tc.color, patch, "keywords/$MailFlagBit1", tc.bit1)
		checkBit(t, tc.color, patch, "keywords/$MailFlagBit2", tc.bit2)
	}
}

func checkBit(t *testing.T, color FlagColor, patch jmap.Patch, key string, wantSet bool) {
	t.Helper()
	v, ok := patch[key]
	if !ok {
		t.Fatalf("FlagColor(%d).Patch() missing %s", color, key)
	}
	if wantSet {
		b, ok := v.(bool)
		if !ok || !b {
			t.Errorf("FlagColor(%d).Patch()[%s] = %#v, want true", color, key, v)
		}
	} else {
		if v != nil {
			t.Errorf("FlagColor(%d).Patch()[%s] = %#v, want nil", color, key, v)
		}
	}
}

// --- clearColorPatch tests ---

func TestClearColorPatch(t *testing.T) {
	patch := clearColorPatch()
	if len(patch) != 3 {
		t.Fatalf("clearColorPatch() has %d keys, want 3", len(patch))
	}
	for _, key := range []string{"keywords/$MailFlagBit0", "keywords/$MailFlagBit1", "keywords/$MailFlagBit2"} {
		v, ok := patch[key]
		if !ok {
			t.Fatalf("clearColorPatch() missing %s", key)
		}
		if v != nil {
			t.Errorf("clearColorPatch()[%s] = %#v, want nil", key, v)
		}
	}
}

// --- SetFlaggedWithColor tests ---

func TestSetFlaggedWithColor_PatchStructure(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			setReq := req.Calls[0].Args.(*email.Set)
			updated := make(map[jmap.ID]*email.Email, len(setReq.Update))
			for id := range setReq.Update {
				updated[id] = &email.Email{}
			}
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/set", CallID: "0", Args: &email.SetResponse{Updated: updated}},
			}}, nil
		},
	}

	succeeded, errs := c.SetFlaggedWithColor([]string{"M1"}, FlagColorOrange)
	if len(succeeded) != 1 {
		t.Fatalf("expected 1 succeeded, got %d", len(succeeded))
	}
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errs))
	}

	setReq := captured.Calls[0].Args.(*email.Set)
	patch := setReq.Update["M1"]

	// Must include $flagged = true.
	v, ok := patch["keywords/$flagged"]
	if !ok {
		t.Fatal("expected keywords/$flagged in patch")
	}
	if flagged, ok := v.(bool); !ok || !flagged {
		t.Fatalf("expected keywords/$flagged=true, got %#v", v)
	}

	// Must include orange color bits: Bit0=true, Bit1=nil, Bit2=nil.
	if v := patch["keywords/$MailFlagBit0"]; v != true {
		t.Errorf("expected $MailFlagBit0=true for orange, got %#v", v)
	}
	if v := patch["keywords/$MailFlagBit1"]; v != nil {
		t.Errorf("expected $MailFlagBit1=nil for orange, got %#v", v)
	}
	if v := patch["keywords/$MailFlagBit2"]; v != nil {
		t.Errorf("expected $MailFlagBit2=nil for orange, got %#v", v)
	}
}

func TestSetFlaggedWithColor_RedClearsAllBits(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			setReq := req.Calls[0].Args.(*email.Set)
			updated := make(map[jmap.ID]*email.Email, len(setReq.Update))
			for id := range setReq.Update {
				updated[id] = &email.Email{}
			}
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/set", CallID: "0", Args: &email.SetResponse{Updated: updated}},
			}}, nil
		},
	}

	c.SetFlaggedWithColor([]string{"M1"}, FlagColorRed)

	setReq := captured.Calls[0].Args.(*email.Set)
	patch := setReq.Update["M1"]

	// Red = 000, all bits should be nil (cleared).
	for _, key := range []string{"keywords/$MailFlagBit0", "keywords/$MailFlagBit1", "keywords/$MailFlagBit2"} {
		if v := patch[key]; v != nil {
			t.Errorf("expected %s=nil for red, got %#v", key, v)
		}
	}
}

// --- Updated SetUnflagged tests ---

func TestSetUnflagged_ClearsColorBits(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			setReq := req.Calls[0].Args.(*email.Set)
			updated := make(map[jmap.ID]*email.Email, len(setReq.Update))
			for id := range setReq.Update {
				updated[id] = &email.Email{}
			}
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/set", CallID: "0", Args: &email.SetResponse{Updated: updated}},
			}}, nil
		},
	}

	c.SetUnflagged([]string{"M1"})

	setReq := captured.Calls[0].Args.(*email.Set)
	patch := setReq.Update["M1"]

	// Must clear $flagged.
	if v := patch["keywords/$flagged"]; v != nil {
		t.Errorf("expected keywords/$flagged=nil, got %#v", v)
	}

	// Must clear all color bits.
	for _, key := range []string{"keywords/$MailFlagBit0", "keywords/$MailFlagBit1", "keywords/$MailFlagBit2"} {
		v, ok := patch[key]
		if !ok {
			t.Fatalf("expected %s in unflag patch", key)
		}
		if v != nil {
			t.Errorf("expected %s=nil in unflag patch, got %#v", key, v)
		}
	}
}

// --- ClearFlagColor tests ---

func TestClearFlagColor_OnlyClearsColorBits(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			setReq := req.Calls[0].Args.(*email.Set)
			updated := make(map[jmap.ID]*email.Email, len(setReq.Update))
			for id := range setReq.Update {
				updated[id] = &email.Email{}
			}
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/set", CallID: "0", Args: &email.SetResponse{Updated: updated}},
			}}, nil
		},
	}

	succeeded, errs := c.ClearFlagColor([]string{"M1"})
	if len(succeeded) != 1 {
		t.Fatalf("expected 1 succeeded, got %d", len(succeeded))
	}
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errs))
	}

	setReq := captured.Calls[0].Args.(*email.Set)
	patch := setReq.Update["M1"]

	// Must NOT include $flagged.
	if _, ok := patch["keywords/$flagged"]; ok {
		t.Error("ClearFlagColor should not touch keywords/$flagged")
	}

	// Must clear all color bits.
	for _, key := range []string{"keywords/$MailFlagBit0", "keywords/$MailFlagBit1", "keywords/$MailFlagBit2"} {
		v, ok := patch[key]
		if !ok {
			t.Fatalf("expected %s in patch", key)
		}
		if v != nil {
			t.Errorf("expected %s=nil, got %#v", key, v)
		}
	}
}

// --- ValidColorNames tests ---

func TestValidColorNames(t *testing.T) {
	names := ValidColorNames()
	if len(names) != 7 {
		t.Fatalf("expected 7 color names, got %d", len(names))
	}
	expected := []string{"red", "orange", "yellow", "green", "blue", "purple", "gray"}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("ValidColorNames()[%d] = %q, want %q", i, names[i], name)
		}
	}
}
