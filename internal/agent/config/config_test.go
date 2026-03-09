package config

import "testing"

func TestPickValue_PriorityOrder(t *testing.T) {
	def := 10
	env := 1
	flag := 2
	json := 3

	tests := []struct {
		name string
		env  *int
		flag *int
		json *int
		want int
	}{
		{
			name: "env has highest priority",
			env:  &env,
			flag: &flag,
			json: &json,
			want: env,
		},
		{
			name: "flag when env is nil",
			env:  nil,
			flag: &flag,
			json: &json,
			want: flag,
		},
		{
			name: "json when env and flag are nil",
			env:  nil,
			flag: nil,
			json: &json,
			want: json,
		},
		{
			name: "default when all zero or nil",
			env:  nil,
			flag: nil,
			json: nil,
			want: def,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pickValue(tt.env, tt.flag, tt.json, def); got != tt.want {
				t.Errorf("pickValue() = %d, want %d", got, tt.want)
			}
		})
	}
}


