package daifugo

import (
	"testing"
)

func Test_hoge(t *testing.T) {
	type args = string
	tests := []struct {
		name string
		args args
		want string
	}{
		{"pass", `
		{
			"type": "pass",
			"data": {
			}
		}`, "pass"},
		{"pass", `
		{
			"type": "submitCard",
			"data": {
				"cards": ["3S", "4D"]
			}
		}`, "submitCard"},
	}
	/*
	for _, tt := range tests {
		
		t.Run(tt.name, func(t *testing.T) {
			if got, reason := parseMessageTypeAndPlayerName(tt.args); got != tt.want {
				t.Errorf("%s failed with reason:'%s'. got %v, want %v", tt.name, reason, got, tt.want)
			}
		})
			
	}
		*/
}