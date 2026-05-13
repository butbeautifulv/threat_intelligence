package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	calderaOwner = "mitre"
	calderaRepo  = "stockpile"
)

func (u *Ingestor) ingestCalderaAbilities(ctx context.Context) error {
	base := strings.TrimSpace(os.Getenv("DS_CALDERA_BASE_PATH"))
	if base == "" {
		base = "data/abilities"
	}
	maxN := envInt("DS_MAX_CALDERA", 200)
	u.logger.Info("ingesting Caldera abilities (Stockpile)", slog.String("path", base), slog.Int("limit", maxN))

	cats, err := u.githubListDir(ctx, calderaOwner, calderaRepo, base)
	if err != nil {
		return fmt.Errorf("caldera list %s: %w", base, err)
	}
	n := 0
	for _, cat := range cats {
		if n >= maxN {
			break
		}
		if cat.Type != "dir" || cat.Name == ".github" {
			continue
		}
		items, err := u.githubListDir(ctx, calderaOwner, calderaRepo, cat.Path)
		if err != nil {
			u.logger.Debug("caldera category skip", slog.String("path", cat.Path), slog.String("err", err.Error()))
			continue
		}
		for _, it := range items {
			if n >= maxN {
				break
			}
			if it.Type != "file" || (!strings.HasSuffix(it.Name, ".yml") && !strings.HasSuffix(it.Name, ".yaml")) {
				continue
			}
			raw, err := u.fetchBytes(ctx, it.DownloadURL, filepath.Join(u.cache, "caldera", cat.Name, it.Name))
			if err != nil {
				continue
			}
			added := u.parseCalderaYAMLFile(ctx, raw, it.Path, &n, maxN)
			if added > 0 {
				u.logger.Debug("caldera file", slog.String("path", it.Path), slog.Int("abilities", added))
			}
		}
	}
	u.logger.Info("caldera ingest done", slog.Int("count", n))
	return nil
}

func (u *Ingestor) parseCalderaYAMLFile(ctx context.Context, raw []byte, path string, counter *int, maxN int) int {
	added := 0
	// Stockpile stores a YAML sequence of ability objects.
	var seq []map[string]any
	if err := yaml.Unmarshal(raw, &seq); err == nil && len(seq) > 0 {
		for _, root := range seq {
			if *counter >= maxN {
				break
			}
			if one := u.tryUpsertCalderaRoot(ctx, root, path); one {
				added++
				*counter++
			}
		}
		return added
	}
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return 0
	}
	if u.tryUpsertCalderaRoot(ctx, root, path) {
		added++
		*counter++
	}
	return added
}

func (u *Ingestor) tryUpsertCalderaRoot(ctx context.Context, root map[string]any, path string) bool {
	id, _ := root["id"].(string)
	name, _ := root["name"].(string)
	desc, _ := root["description"].(string)
	tactic, _ := root["tactic"].(string)
	techID := ""
	if tm, ok := root["technique"].(map[string]any); ok {
		techID, _ = tm["attack_id"].(string)
	}
	if id == "" {
		return false
	}
	md := fmt.Sprintf("# %s\n\n**Tactic:** %s  \n**Technique:** %s  \n\n%s\n", name, tactic, techID, desc)
	if err := u.store.UpsertCalderaAbility(ctx, id, name, tactic, techID, md, "mitre-stockpile"); err != nil {
		u.logger.Warn("caldera upsert failed", slog.String("id", id), slog.String("err", err.Error()))
		return false
	}
	return true
}
