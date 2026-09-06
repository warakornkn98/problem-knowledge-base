package slug

import "testing"

func TestMake(t *testing.T) {
	cases := map[string]string{
		"Backend":           "backend",
		"  Reverse Proxy  ": "reverse-proxy",
		"CI/CD Pipeline":    "ci-cd-pipeline",
		"PostgreSQL 16":     "postgresql-16",
		"api--gateway":      "api-gateway",
		"เครือข่าย Network": "เครือข่าย-network",
		"!!!":               "",
	}
	for in, want := range cases {
		if got := Make(in); got != want {
			t.Errorf("Make(%q) = %q, want %q", in, got, want)
		}
	}
}
