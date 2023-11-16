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
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/manifoldco/promptui"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

const githubURL = "https://github.com/"

var lang, spokenLang, period, mode string

// var browser bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gh-trending",
	Short: "Check the Github Trending(https://github.com/trending) in the TUI and navigate to the repository page.",
	RunE: func(cmd *cobra.Command, args []string) error {
		query := url.Values{
			"spoken_language_code": []string{spokenLang},
			"since":                []string{period},
		}
		u := &url.URL{
			Scheme:   "https",
			Host:     "github.com",
			Path:     path.Join("trending", strings.ToLower(lang)),
			RawQuery: query.Encode(),
		}
		resp, err := http.Get(u.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "HTTP Request/Response error: %v", err)
			return err
		}
		defer resp.Body.Close()
		var r io.Reader = resp.Body
		repos, err := ParseRepos(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "HTML Parse error: %v", err)
			return err
		}
		if len(repos) == 0 {
			fmt.Fprintln(os.Stdout, "No Results.")
			return nil
		}

		if mode == "json" {
			jsonBytes, err := json.Marshal(repos)
			if err != nil {
				fmt.Fprintf(os.Stderr, "JSON Marshal error: %v\n", err)
				return err
			}
			var buf bytes.Buffer
			json.Indent(&buf, jsonBytes, "", "    ")
			fmt.Fprintln(os.Stdout, buf.String())
			return nil
		}

		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "\U0001f449 {{ .RepoName }}",
			Inactive: "    {{ .RepoName }}",
			Selected: "\U0001F336  {{ .RepoName }}",
			Details: `
	--------- Repository Details----------
	{{ "RepoName:" | faint }}	{{ .RepoName }}
	{{ "Description:" | faint }}	{{ .Description }}
	{{ "Language:" | faint }}	{{ .Language }}
	{{ "Stars:" | faint }}	{{ .Stars }}
	{{ "StarsInPeriod:" | faint }}	{{ .StarsInPeriod }}
	{{ "Forks:" | faint }}	{{ .Forks }}
	`,
		}

		searcher := func(input string, index int) bool {
			repo := repos[index]
			name := strings.Replace(strings.ToLower(repo.RepoName), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label:     "Selecting a repository opens a page for that repository.",
			Items:     repos,
			Templates: templates,
			Size:      4,
			Searcher:  searcher,
		}

		i, _, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return err
		}
		browser.OpenURL(repos[i].Href)
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Option for creating access URL
	rootCmd.Flags().StringVarP(&lang, "language", "l", "", "Programming Language: go, typescript, ruby, .... anything is ok!")
	rootCmd.Flags().StringVarP(&spokenLang, "spoken-language", "s", "", "Spoken Language: en(English), zh(Chinese), ja(Japanese), and so on.")
	rootCmd.Flags().StringVarP(&period, "period", "p", "today", "Date Range: today, weekly or monthly")
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "browser", "Startup mode: browser(Select a trend repository and open its Github page) or json")
}

func ParseRepos(r io.Reader) ([]Repository, error) {
	repos := []Repository{}
	doc, err := htmlquery.Parse(r)
	if err != nil {
		return nil, err
	}

	var repo *Repository
	articleList := htmlquery.Find(doc, "//article")
	for _, article := range articleList {
		repo = &Repository{}
		// Href, RepoName
		repoLinkNode := htmlquery.FindOne(article, "/h2/a")
		if repoLinkNode == nil {
			return nil, fmt.Errorf("repository link cannot be found")
		}

		hrefPath := htmlquery.SelectAttr(repoLinkNode, "href")
		repo.Href = path.Join(githubURL, hrefPath)
		repoName := strings.Split(hrefPath, "/")
		repo.RepoName = repoName[1] + "/" + repoName[2]

		// Description
		descNode, err := htmlquery.QueryAll(article, "/p")
		if descNode != nil && err == nil {
			repo.Description = strings.TrimSpace(htmlquery.InnerText(descNode[0]))
		}

		// Language
		langNode, err := htmlquery.QueryAll(article, "/div[2]/span[1]/span[2]")
		if langNode != nil && err == nil {
			repo.Language = htmlquery.InnerText(langNode[0])
		}

		// Stars
		starsNode, err := htmlquery.QueryAll(article, "/div[2]/a[1]")
		if starsNode != nil && err == nil {
			repo.Stars, err = strconv.Atoi(strings.TrimSpace(strings.Replace(htmlquery.InnerText(starsNode[0]), ",", "", -1)))
			if err != nil {
				log.Fatal(err)
			}
		}

		// Forks
		forksNode, err := htmlquery.QueryAll(article, "/div[2]/a[2]")
		if forksNode != nil && err == nil {
			repo.Forks, err = strconv.Atoi(strings.TrimSpace(strings.Replace(htmlquery.InnerText(forksNode[0]), ",", "", -1)))
			if err != nil {
				log.Fatal(err)
			}
		}

		// StarsInPeriod
		starsInPeriodNode, err := htmlquery.QueryAll(article, "/div[2]/span[3]")
		if starsInPeriodNode != nil && err == nil {
			repo.StarsInPeriod = strings.TrimSpace(htmlquery.InnerText(starsInPeriodNode[0]))
		}
		repos = append(repos, *repo)
	}
	return repos, nil
}

type Repository struct {
	RepoName      string `json:"repoName"`
	Href          string `json:"href"`
	Description   string `json:"description"`
	Language      string `json:"language"`
	Stars         int    `json:"stars"`
	Forks         int    `json:"forks"`
	StarsInPeriod string `json:"starsInPeriod"`
}
