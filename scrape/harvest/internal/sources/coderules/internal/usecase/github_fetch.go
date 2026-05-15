package usecase

import (
	"context"
	"net/http"

	"github.com/butbeautifulv/threat_intelligence/scrape/pkg/githubraw"
	"github.com/butbeautifulv/threat_intelligence/scrape/harvest/internal/feeds"
	"github.com/butbeautifulv/threat_intelligence/scrape/harvest/internal/ledger"

	gh "github.com/butbeautifulv/threat_intelligence/scrape/harvest/internal/sources/coderules/internal/feeds/github"
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
	var ledgerFn githubraw.FetchIfDueFunc
	if r.ledger != nil {
		ledgerFn = func(ctx context.Context, key, source, url, cachePath string, buildReq func() (*http.Request, error)) (githubraw.FetchResult, error) {
			res, err := feeds.FetchIfDue(ctx, fc, r.ledger, key, source, url, ledger.PolicyPeriodic, cachePath, buildReq)
			return githubraw.FetchResult{Body: res.Body, Skipped: res.Skipped, Unchanged: res.Unchanged}, err
		}
	}
	return githubraw.FetchFile(githubraw.FileParams{
		Ctx: ctx, Owner: owner, Repo: repo, Ref: ref, Path: ghPath,
		RawURL: rawURL, CacheRel: cacheRel, LedgerKey: "gh:file:" + owner + ":" + repo + ":" + ghPath, Source: "coderules",
		UserAgent: "veil-scrape/1.0", Ledger: ledgerFn,
		FetchRaw: func(ctx context.Context, o, rep, rf, p string) ([]byte, error) {
			return feeds.GitHubFetchRaw(ctx, fc, o, rep, rf, p)
		},
		SkipErrFmt: "download %s skipped without cache",
	})
}
