package transmission

import (
	"net/http"
	"strings"
)

func httpClient() *http.Client {
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	return &client
}

func matchCategory(input []string, categories map[string]struct {
	Path    string `yaml:"path"`
	Matcher string `yaml:"matcher"`
},
) string {
	for i, j := range categories {
		matchers := strings.Split(j.Matcher, ",")
		for _, k := range input {
			for _, l := range matchers {
				if k == l {
					return i
				}
			}
		}
	}
	return "noop"
}
