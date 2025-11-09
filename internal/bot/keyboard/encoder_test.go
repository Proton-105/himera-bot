package keyboard_test

import (
	"strings"
	"testing"

	"github.com/Proton-105/himera-bot/internal/bot/keyboard"
)

func TestEncodeCallback(t *testing.T) {
	tests := []struct {
		name      string
		unique    string
		data      string
		want      string
		wantError bool
	}{
		{
			name:   "with data",
			unique: "history",
			data:   "2",
			want:   "history:2",
		},
		{
			name:   "without data",
			unique: "tokens",
			data:   "",
			want:   "tokens",
		},
		{
			name:      "exceeds limit",
			unique:    strings.Repeat("x", keyboard.CallbackDataLimitBytes+1),
			data:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := keyboard.EncodeCallback(tt.unique, tt.data)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("EncodeCallback() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDecodeCallback(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantUnique string
		wantData   string
		wantErr    bool
	}{
		{
			name:       "unique and data",
			input:      "history:3",
			wantUnique: "history",
			wantData:   "3",
		},
		{
			name:       "only unique",
			input:      "tokens",
			wantUnique: "tokens",
			wantData:   "",
		},
		{
			name:       "multiple separators",
			input:      "action:part1:part2",
			wantUnique: "action",
			wantData:   "part1:part2",
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unique, data, err := keyboard.DecodeCallback(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if unique != tt.wantUnique || data != tt.wantData {
				t.Errorf("DecodeCallback() = (%q, %q), want (%q, %q)", unique, data, tt.wantUnique, tt.wantData)
			}
		})
	}
}
