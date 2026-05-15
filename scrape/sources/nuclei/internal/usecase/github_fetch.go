package usecase

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/scrape/feeds"
	"github.com/butbeautifulv/threat_intelligence/scrape/ledger"

	gh "github.com/butbeautifulv/threat_intelligence/scrape/sources/nuclei/internal/feeds/github"
)

func nucleiGitRef() string {
	return "master"
}

func (r *Runner) githubListDir(ctx context.Context, owner, repo, path string) ([]gh.Content, error) {
	fc := r.feeds
	if fc == nil {
		fc = feeds.NewClient("", r.log)
	}
	return gh.NewClient(fc).ListDir(ctx, owner, repo, path)
}

func (r *Runner) fetchTemplateFile(ctx context.Context, owner, repo, ghPath string) ([]byte, bool, error) {
	ref := nucleiGitRef()
	rawURL := feeds.GitHubRawURL(owner, repo, ref, ghPath)
	cacheRel := filepath.Join("nuclei", strings.ReplaceAll(ghPath, "/", "__"))
	fc := r.feeds
	if fc == nil {
		fc = feeds.NewClient("", r.log)
	}
	if r.ledger == nil {
		raw, err := feeds.GitHubFetchRaw(ctx, fc, owner, repo, ref, ghPath)
		return raw, false, err
	}
	key := fmt.Sprintf("nuclei:file:%s", ghPath)
	res, err := feeds.FetchIfDue(ctx, fc, r.ledger, key, "nuclei", rawURL, ledger.PolicyPeriodic, cacheRel, func() (*http.Request, error) {
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
		return nil, false, fmt.Errorf("nuclei file %s skipped without cache", ghPath)
	}
	return res.Body, false, nil
}
