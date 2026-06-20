package ai

import "testing"

func TestParseModel(t *testing.T) {
	cases := []struct {
		ref  string
		want string
		caps ModelCapabilities
	}{
		{"gpt-4o", "gpt-4o", ModelCapabilities{Chat: true}},
		{"openai:gpt-4o", "gpt-4o", ModelCapabilities{Chat: true, Tools: true, JSONMode: true}},
		{"anthropic:claude-3-5-sonnet", "claude-3-5-sonnet", ModelCapabilities{Chat: true, Tools: true}},
		{"custom:model", "model", ModelCapabilities{Chat: true}},
	}

	for _, tc := range cases {
		m, err := ParseModel(tc.ref)
		if err != nil {
			t.Fatalf("ParseModel(%q): %v", tc.ref, err)
		}
		if m.ID() != tc.want {
			t.Fatalf("ParseModel(%q) id = %q, want %q", tc.ref, m.ID(), tc.want)
		}
		caps := m.Capabilities()
		if caps != tc.caps {
			t.Fatalf("ParseModel(%q) caps = %+v, want %+v", tc.ref, caps, tc.caps)
		}
	}
}

func TestParseModelEmpty(t *testing.T) {
	_, err := ParseModel("")
	if err == nil {
		t.Fatal("expected error for empty model ref")
	}
}

func TestModelCapabilitiesHas(t *testing.T) {
	caps := ModelCapabilities{Chat: true, Tools: true}
	if !caps.Has(ModelCapabilities{Chat: true}) {
		t.Fatal("expected Chat capability")
	}
	if caps.Has(ModelCapabilities{Embedding: true}) {
		t.Fatal("did not expect Embedding capability")
	}
}
