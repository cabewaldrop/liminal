package matcher

import (
	"fmt"
	"net/http"
	"strings"
)

type Matcher interface {
	Match(r *http.Request) (string, error)
}

func NewMatcher(strategy string) Matcher {
	if strategy == "Host" {
		return HostMatcher{}
	}

	return IPMatcher{}
}

type IPMatcher struct{}

func (m IPMatcher) Match(r *http.Request) (string, error) {
	parsed := strings.SplitN(r.RemoteAddr, ":", 2)
	if parsed[0] == "" {
		return "", fmt.Errorf("Unable to parse ip address from: %s", r.RemoteAddr)
	}

	return parsed[0], nil
}

type HostMatcher struct{}

func (m HostMatcher) Match(r *http.Request) (string, error) {
	return r.Host, nil
}
