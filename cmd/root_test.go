/*
Copyright © 2022 Takafumi Umemoto <takafumi.umemoto@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd_test

import (
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/nmemoto/gh-trending/cmd"
)

// Verify that the current Trends Page can properly fetch information
func TestParseRepos(t *testing.T) {
	// Check behavior in default state
	u := &url.URL{
		Scheme:   "https",
		Host:     "github.com",
		Path:     path.Join("trending", ""),
		RawQuery: "",
	}

	resp, err := http.Get(u.String())
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	defer resp.Body.Close()

	var r io.Reader = resp.Body
	repos, err := cmd.ParseRepos(r)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// If "repos" is 0, it ends normally as "No Results".
	if len(repos) == 0 {
		return
	}

	for _, repo := range repos {
		// Check href
		url, err := url.Parse(repo.Href)
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		if url.Scheme != "https" || url.Host != "github.com" {
			t.Errorf("href is not valid: %s, expected, expect 'https' and 'github.com 'to be used for value", repo.Href)
		}
		// expected to be like "github.com","owner","repo"
		if len(strings.Split(url.Path, "/")) != 3 {
			t.Errorf("href is not valid: %s", repo.Href)
		}

		// Check repoName
		if !regexp.MustCompile(`^[\w-]+/[\w-]+$`).MatchString(repo.RepoName) {
			t.Errorf("repoName is not valid: %s, expected to be like owner/repo", repo.RepoName)
		}
		if !strings.Contains(repo.Href, repo.RepoName) {
			t.Errorf("repoName should be included in href: %s, %s", repo.Href, repo.RepoName)
		}

		// TODO: Other repository members
		// // Check starsInPeriod, expected to be like 1,234 stars today, or 1,234 stars this week etc...
		// if repo.StarsInPeriod == "" {
		// 	t.Errorf("%s, starsInPeriod is empty", repo.RepoName)
		// }
		// if !regexp.MustCompile(`^\d+ stars (today|this week|this month)$`).MatchString(repo.StarsInPeriod) {
		// 	t.Errorf("%s, starsInPeriod is not valid: %s", repo.RepoName, repo.StarsInPeriod)
		// }
	}
}
