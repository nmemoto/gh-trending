# gh-trending

Check the Github Trending(https://github.com/trending) in the TUI and navigate to the repository page.

## Install

```bash
$ gh extension install nmemoto/gh-trending
```

## Usage

```bash
$ gh trending -l go
Use the arrow keys to navigate: ↓ ↑ → ←  and / toggles search
Selecting a repository opens a page for that repository.
↑ 👉 external-secrets/external-secrets
      sealerio/sealer
      tmrts/go-patterns
↓     gogf/gf

        --------- Repository Details----------
        RepoName:             external-secrets/external-secrets
        Description:          External Secrets Operator reads information from a third-party service like AWS Secrets Manager and automatically injects the values as Kubernetes Secrets.
        Language:             Go
        Stars:                1054
        StarsInPeriod:        10 stars today
        Forks:                204
```

```bash
$ gh trending --help
Check the Github Trending(https://github.com/trending) in the TUI and navigate to the repository page.

Usage:
  gh-trending [flags]

Flags:
  -h, --help                     help for gh-trending
  -l, --language string          Programming Language: go, typescript, ruby, .... anything is ok!
  -m, --mode string              Startup mode: browser(Select a trend repository and open its Github page) or json (default "browser")
  -p, --period string            Date Range: today, weekly or monthly (default "today")
  -s, --spoken-language string   Spoken Language: en(English), zh(Chinese), ja(Japanese), and so on.
```

