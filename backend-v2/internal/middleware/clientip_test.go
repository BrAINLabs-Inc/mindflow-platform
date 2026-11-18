package middleware

import "testing"

// Expected values follow proxy-addr's (Express `trust proxy = N`) rules.
func TestResolveClientIP(t *testing.T) {
	cases := []struct {
		name, remote, xff string
		hops              int
		want              string
	}{
		{"no proxy trust ignores XFF", "10.0.0.1:5000", "1.2.3.4", 0, "10.0.0.1"},
		{"no XFF uses socket", "10.0.0.1:5000", "", 1, "10.0.0.1"},
		{"one hop takes right-most XFF", "10.0.0.1:5000", "6.6.6.6, 1.2.3.4", 1, "1.2.3.4"},
		{"spoofed left entries are ignored", "10.0.0.1:5000", "9.9.9.9, 8.8.8.8, 1.2.3.4", 1, "1.2.3.4"},
		{"two hops", "10.0.0.1:5000", "6.6.6.6, 1.2.3.4, 10.0.0.2", 2, "1.2.3.4"},
		{"more hops than entries clamps to left-most", "10.0.0.1:5000", "1.2.3.4", 5, "1.2.3.4"},
		{"remote without port", "10.0.0.1", "", 1, "10.0.0.1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveClientIP(tc.remote, tc.xff, tc.hops); got != tc.want {
				t.Errorf("resolveClientIP(%q, %q, %d) = %q, want %q", tc.remote, tc.xff, tc.hops, got, tc.want)
			}
		})
	}
}
