package buildinfo

import "testing"

func TestLine(t *testing.T) {
	defer func(v, c, d string) { Version, Commit, Date = v, c, d }(Version, Commit, Date)

	tests := []struct {
		name    string
		version string
		commit  string
		date    string
		want    string
	}{
		{"dev build", "dev", "", "", "md dev"},
		{"release", "v1.2.3", "abc1234", "2026-07-26T12:00:00Z",
			"md v1.2.3 (commit abc1234, built 2026-07-26T12:00:00Z)"},
		{"version and commit", "v1.2.3", "abc1234", "", "md v1.2.3 (commit abc1234)"},
		{"version and date", "v1.2.3", "", "2026-07-26T12:00:00Z",
			"md v1.2.3 (built 2026-07-26T12:00:00Z)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version, Commit, Date = tt.version, tt.commit, tt.date
			if got := Line("md"); got != tt.want {
				t.Errorf("Line(md) = %q, want %q", got, tt.want)
			}
		})
	}
}
