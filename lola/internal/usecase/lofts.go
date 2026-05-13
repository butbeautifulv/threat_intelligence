package usecase

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

func (u *ScraperUsecase) IngestLOFTS(ctx context.Context) error {
	url := strings.TrimSpace(os.Getenv("LOFTS_URL"))
	if url == "" {
		url = "https://lofts.galeal.com/"
	}
	skip := strings.EqualFold(strings.TrimSpace(os.Getenv("LOFTS_SKIP_ON_ERROR")), "true")
	u.logger.Info("ingesting LOFTS", slog.String("url", url))

	cacheFile := filepath.Join(u.cache, "lofts", "index.html")
	var body []byte
	if u.cache != "" {
		if b, err := os.ReadFile(cacheFile); err == nil && len(b) > 100 {
			body = b
		}
	}
	if len(body) == 0 {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "threat_intelligence-lola/1.0")
		resp, err := u.http.Do(req)
		if err != nil {
			if skip {
				u.logger.Warn("lofts fetch failed; skipping", slog.String("err", err.Error()))
				return nil
			}
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			if skip {
				u.logger.Warn("lofts http error; skipping", slog.Int("code", resp.StatusCode), slog.String("body", string(b)))
				return nil
			}
			return fmt.Errorf("lofts http %d: %s", resp.StatusCode, string(b))
		}
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			if skip {
				u.logger.Warn("lofts read failed; skipping", slog.String("err", err.Error()))
				return nil
			}
			return err
		}
		if u.cache != "" {
			_ = os.MkdirAll(filepath.Dir(cacheFile), 0o755)
			_ = os.WriteFile(cacheFile, body, 0o644)
		}
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	entries := extractLoftsLinks(doc)
	maxE := 200
	if v := os.Getenv("LOFTS_MAX_ENTRIES"); v != "" {
		var n int
		_, _ = fmt.Sscanf(v, "%d", &n)
		if n > 0 {
			maxE = n
		}
	}
	n := 0
	for _, e := range entries {
		if n >= maxE {
			break
		}
		if e.href == "" || strings.HasPrefix(e.href, "#") {
			continue
		}
		link := e.href
		if strings.HasPrefix(link, "/") {
			link = strings.TrimRight(url, "/") + link
		}
		md := fmt.Sprintf("# %s\n\n**Link:** %s\n", e.text, link)
		if err := u.repo.UpsertLoftsEntry(ctx, e.text, e.category, link, md); err != nil {
			return err
		}
		n++
	}
	u.logger.Info("LOFTS ingest done", slog.Int("entries", n))
	return nil
}

type loftsLink struct {
	href, text, category string
}

func extractLoftsLinks(n *html.Node) []loftsLink {
	var out []loftsLink
	var walk func(*html.Node, string)
	walk = func(node *html.Node, category string) {
		if node.Type == html.ElementNode && node.Data == "h2" {
			category = textContent(node)
		}
		if node.Type == html.ElementNode && node.Data == "a" {
			var href string
			for _, a := range node.Attr {
				if a.Key == "href" {
					href = strings.TrimSpace(a.Val)
					break
				}
			}
			if href != "" {
				t := strings.TrimSpace(textContent(node))
				if t != "" {
					out = append(out, loftsLink{href: href, text: t, category: category})
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c, category)
		}
	}
	walk(n, "")
	return out
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(b.String())
}
