package sanitize

import "strings"

func Text(input string, secrets ...string) string {
	out := input
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		out = strings.ReplaceAll(out, secret, "[REDACTED]")
	}
	return out
}

func Path(input string) string {
	for _, marker := range []string{"/Users/", "/home/", "/root/"} {
		idx := strings.Index(input, marker)
		if idx >= 0 {
			rest := input[idx+len(marker):]
			parts := strings.Split(rest, "/")
			if len(parts) > 0 && parts[0] != "" {
				return strings.Replace(input, marker+parts[0], marker+"[user]", 1)
			}
		}
	}
	return input
}
