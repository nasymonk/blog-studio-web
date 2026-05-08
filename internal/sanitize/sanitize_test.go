package sanitize

import "testing"

func TestText(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		secrets []string
		want    string
	}{
		{"no secrets", "hello world", nil, "hello world"},
		{"single secret", "password is abc123", []string{"abc123"}, "password is [REDACTED]"},
		{"multiple secrets", "a=foo b=bar", []string{"foo", "bar"}, "a=[REDACTED] b=[REDACTED]"},
		{"empty secret skipped", "data: secret", []string{"", "secret"}, "data: [REDACTED]"},
		{"secret not present", "nothing here", []string{"other"}, "nothing here"},
		{"secret appears multiple times", "x=s x=s", []string{"s"}, "x=[REDACTED] x=[REDACTED]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Text(tt.input, tt.secrets...)
			if got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/Users/alice/projects/foo", "/Users/[user]/projects/foo"},
		{"/home/bob/stuff", "/home/[user]/stuff"},
		{"/root/data", "/root/[user]"},
		{"/var/log/app.log", "/var/log/app.log"},
		{"no path at all", "no path at all"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Path(tt.input)
			if got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}
