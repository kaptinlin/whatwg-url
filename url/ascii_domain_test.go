package url

import (
	"testing"

	"github.com/nlnwa/whatwg-url/errors"
)

// The ASCII domain rule is independent of Unicode ToASCII's outcome:
// https://url.spec.whatwg.org/commit-snapshots/845bafca22fb0c5a42c4bf384a9bdea525a2f10d/#concept-domain-to-ascii
// The xn-- URL fixtures in testdata/urltestdata.json are copied from WPT:
// https://github.com/web-platform-tests/wpt/blob/647d3bdf133159739b57cfb7afa0be3f5d76b9db/url/resources/urltestdata.json
func TestParseASCIIDomain(t *testing.T) {
	tests := []struct {
		input string
		href  string
	}{
		{"https://xn--/", "https://xn--/"},
		{"https://XN--/", "https://xn--/"},
		{"https://xn--8i7caa/", "https://xn--8i7caa/"},
		{"https://xn--a/", "https://xn--a/"},
		{"https://%78%6E--/", "https://xn--/"},
		{"file://XN--/p", "file://xn--/p"},
		{"https://xn--.example/", "https://xn--.example/"},
		{"https://xn--bcher-kva.example/", "https://xn--bcher-kva.example/"},
		{"https://bücher.example/", "https://xn--bcher-kva.example/"},
		{"https://127.1/", "https://127.0.0.1/"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			u, err := Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if got := u.Href(false); got != tt.href {
				t.Fatalf("Href() = %q, want %q", got, tt.href)
			}
			roundtrip, err := Parse(u.Href(false))
			if err != nil || roundtrip.Href(false) != tt.href {
				t.Fatalf("roundtrip = %v, %v; want %q", roundtrip, err, tt.href)
			}
		})
	}
}

func TestParseASCIIDomainRejectsInvalidHosts(t *testing.T) {
	for _, tt := range []struct {
		input string
		code  errors.ErrorType
	}{
		{"https://xn--é/", errors.DomainToASCII},
		{"https://xn--a.ß/", errors.DomainToASCII},
		{"https://xn--%20/", errors.DomainInvalidCodePoint},
		{"https://xn--%2F/", errors.DomainInvalidCodePoint},
		{"https://xn--%00/", errors.DomainInvalidCodePoint},
		{"https://xn--%25/", errors.DomainInvalidCodePoint},
		{"https://xn--%FF/", errors.DomainToASCII},
		{"https://xn--.1/", errors.IPv4NonNumericPart},
		{"https://example.255/", errors.IPv4NonNumericPart},
		{"https://256.0.0.1/", errors.IPv4OutOfRangePart},
	} {
		t.Run(tt.input, func(t *testing.T) {
			if _, err := Parse(tt.input); err == nil || errors.Type(err) != tt.code {
				t.Fatalf("Parse() error = %v, want %v", err, tt.code)
			}
		})
	}
}

func TestASCIIDomainValidationErrors(t *testing.T) {
	p := NewParser(WithReportValidationErrors())
	u, err := p.Parse("https://xn--/")
	if err != nil {
		t.Fatal(err)
	}
	validationErrors := u.ValidationErrors()
	if len(validationErrors) != 1 || errors.Type(validationErrors[0]) != errors.DomainToASCII || errors.Failure(validationErrors[0]) {
		t.Fatalf("ValidationErrors() = %v, want a non-fatal domain-to-ASCII error", validationErrors)
	}

	p = NewParser(WithFailOnValidationError())
	if _, err := p.Parse("https://xn--/"); err == nil || errors.Type(err) != errors.DomainToASCII {
		t.Fatalf("Parse() error = %v, want domain-to-ASCII with fail-on-validation enabled", err)
	}
}

func TestSetASCIIDomain(t *testing.T) {
	u, err := Parse("https://example.org/path")
	if err != nil {
		t.Fatal(err)
	}
	u.SetHostname("XN--")
	if got := u.Href(false); got != "https://xn--/path" {
		t.Fatalf("Href() = %q, want https://xn--/path", got)
	}
	u.SetHostname("xn--%20")
	if got := u.Href(false); got != "https://xn--/path" {
		t.Fatalf("invalid hostname changed URL to %q", got)
	}
}
