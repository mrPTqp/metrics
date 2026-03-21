package models

import "testing"

func TestNetAddress_SetAddress(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantHost   string
		wantPort   int
		wantString string
	}{
		{
			name:       "valid address",
			input:      "localhost:8080",
			wantErr:    false,
			wantHost:   "localhost",
			wantPort:   8080,
			wantString: "localhost:8080",
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "invalid port",
			input:   "localhost:notaport",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var na NetAddress
			err := na.SetAddress(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("SetAddress(%q) error = nil, want non-nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("SetAddress(%q) unexpected error: %v", tt.input, err)
			}
			if na.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", na.Host, tt.wantHost)
			}
			if na.Port != tt.wantPort {
				t.Errorf("Port = %d, want %d", na.Port, tt.wantPort)
			}
			if got := na.String(); got != tt.wantString {
				t.Errorf("String() = %q, want %q", got, tt.wantString)
			}
		})
	}
}


