package generate

import "testing"

type testCase struct {
	typeName         string
	wantDefaultValue string
	wantOk           bool
}

func TestGetDefautValue(t *testing.T) {
	tests := []testCase{
		{typeName: "int", wantDefaultValue: "0", wantOk: true},
		{typeName: "int64", wantDefaultValue: "0", wantOk: true},
		{typeName: "uint", wantDefaultValue: "0", wantOk: true},
		{typeName: "float64", wantDefaultValue: "0", wantOk: true},
		{typeName: "bool", wantDefaultValue: "false", wantOk: true},
		{typeName: "string", wantDefaultValue: `""`, wantOk: true},
		{typeName: "customType", wantDefaultValue: `""`, wantOk: false},
	}

	for _, testCase := range tests {
		t.Run(testCase.typeName, func(t *testing.T) {
			gotValue, gotOk := getDefautValue(testCase.typeName)

			if gotValue != testCase.wantDefaultValue {
				t.Errorf("getDefautValue(%q) value = %q, want %q",
					testCase.typeName, gotValue, testCase.wantDefaultValue)
			}

			if gotOk != testCase.wantOk {
				t.Errorf("getDefautValue(%q) ok = %v, want %v",
					testCase.typeName, gotOk, testCase.wantOk)
			}
		})
	}
}
