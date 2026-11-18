package protocol

import (
	"strings"
	"testing"
)

func TestParseCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		typeOf  Type
	}{
		{"move", `{"version":1,"type":"move","payload":{"dx":12,"dy":-4}}`, TypeMove},
		{"click", `{"version":1,"type":"click","payload":{"button":"left"}}`, TypeClick},
		{"text", `{"version":1,"type":"text","payload":{"value":"Привет"}}`, TypeText},
		{"key", `{"version":1,"type":"key","payload":{"key":"enter"}}`, TypeKey},
		{"volume", `{"version":1,"type":"volume","payload":{"delta":-5}}`, TypeVolume},
		{"mute", `{"version":1,"type":"mute","payload":{"muted":true}}`, TypeMute},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			command, violation := Parse([]byte(test.message))
			if violation != nil {
				t.Fatalf("Parse() violation = %v", violation)
			}
			if command.Type != test.typeOf {
				t.Fatalf("type = %q, want %q", command.Type, test.typeOf)
			}
		})
	}
}

func TestParseRejectsInvalidPayloadWithoutClosingConnection(t *testing.T) {
	t.Parallel()

	tests := []string{
		`{"version":1,"type":"move","payload":{"dx":201,"dy":0}}`,
		`{"version":1,"type":"click","payload":{"button":"middle"}}`,
		`{"version":1,"type":"text","payload":{"value":"` + strings.Repeat("a", MaxTextBytes+1) + `"}}`,
		`{"version":1,"type":"key","payload":{"key":"tab"}}`,
		`{"version":1,"type":"volume","payload":{"delta":0}}`,
		`{"version":1,"type":"mute","payload":{}}`,
	}

	for _, message := range tests {
		_, violation := Parse([]byte(message))
		if violation == nil || violation.Code != "invalid_payload" || violation.Fatal {
			t.Fatalf("Parse(%q) violation = %#v", message, violation)
		}
	}
}

func TestParseProtocolViolationsAreFatal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		status  int
	}{
		{"malformed", `{`, 1007},
		{"missing payload", `{"version":1,"type":"move"}`, 1007},
		{"unknown envelope field", `{"version":1,"type":"move","payload":{},"extra":true}`, 1007},
		{"unsupported version", `{"version":2,"type":"move","payload":{}}`, 1002},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, violation := Parse([]byte(test.message))
			if violation == nil || !violation.Fatal || violation.CloseStatus != test.status {
				t.Fatalf("violation = %#v, want fatal status %d", violation, test.status)
			}
		})
	}
}

func TestUnknownCommandIsNonFatalAndKeepsID(t *testing.T) {
	t.Parallel()

	command, violation := Parse([]byte(`{"version":1,"type":"dance","id":"command-7","payload":{}}`))
	if violation == nil || violation.Code != "unknown_command" || violation.Fatal {
		t.Fatalf("violation = %#v", violation)
	}
	if command.ID != "command-7" {
		t.Fatalf("id = %q", command.ID)
	}
}
