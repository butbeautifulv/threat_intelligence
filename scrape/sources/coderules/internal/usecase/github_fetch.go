package usecase

import (
	"context"
	"fmt"
	"net/http"

	"github.com/butbeautifulv/threat_intelligence/scrape/feeds"
	"github.com/butbeautifulv/threat_intelligence/scrape/ledger"

	gh "github.com/butbeautifulv/threat_intelligence/scrape/sources/coderules/internal/feeds/github"
)

func coderulesGitRef() string {
	return "master"
}

func (r *Runner) githubListDir(ctx context.Context, owner, repo, path string) ([]gh.Content, error) {
	fc := r.feeds
	if fc == nil {
		fc = feeds.NewClient("", r.log)
	}
	return gh.NewClient(fc).ListDir(ctx, owner, repo, path)
}

func (r *Runner) fetchGitHubFile(ctx context.Context, owner, repo, ghPath, cacheRel string) ([]byte, bool, error) {
	ref := coderulesGitRef()
	rawURL := feeds.GitHubRawURL(owner, repo, ref, ghPath)
	fc := r.feeds
	if fc == nil {
		fc = feeds.NewClient("", r.log)
	}
	if r.ledger == nil {
		raw, err := feeds.GitHubFetchRaw(ctx, fc, owner, repo, ref, ghPath)
		return raw, false, err
	}
	key := fmt.Sprintf("gh:file:%s:%s:%s", owner, repo, ghPath)
	res, err := feeds.FetchIfDue(ctx, fc, r.ledger, key, "coderules", rawURL, ledger.PolicyPeriodic, cacheRel, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "veil-scrape/1.0")
		return req, nil
	})
	if err != nil {
		return nil, false, err
	}
	if res.Unchanged {
		return nil, true, nil
	}
	if res.Skipped && len(res.Body) == 0 {
		return nil, false, fmt.Errorf("download %s skipped without cache", ghPath)
	}
	return res.Body, false, nil
}
