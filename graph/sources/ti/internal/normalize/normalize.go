package normalize

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/graph/sources/ti/internal/domain"
)

var (
	spaceRe   = regexp.MustCompile(`\s+`)
	hexHashRe = regexp.MustCompile(`^[a-f0-9]{32}$|^[a-f0-9]{40}$|^[a-f0-9]{64}$`)
)

// CanonicalID returns a stable ID for IOC node keying.
func CanonicalID(i domain.IOC) string {
	// Keep it deterministic and storage-safe even for long URLs.
	h := sha256.Sum256([]byte(string(i.Type) + ":" + i.Value))
	return hex.EncodeToString(h[:])
}

func NormalizeIOC(i domain.IOC) (domain.IOC, bool) {
	i.Value = strings.TrimSpace(i.Value)
	i.Source = strings.TrimSpace(i.Source)
	var srcs []string
	seen := map[string]struct{}{}
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		srcs = append(srcs, s)
	}
	for _, s := range i.Sources {
		add(s)
	}
	if i.Source != "" {
		add(i.Source)
	}
	sort.Strings(srcs)
	i.Sources = srcs

	switch i.Type {
	case domain.IOCIP:
		ip := net.ParseIP(i.Value)
		if ip == nil {
			return domain.IOC{}, false
		}
		// Normalize to string form.
		i.Value = ip.String()
		return i, true

	case domain.IOCDomain:
		v := strings.ToLower(strings.TrimSuffix(i.Value, "."))
		v = spaceRe.ReplaceAllString(v, "")
		if v == "" || strings.ContainsAny(v, "/:@") {
			return domain.IOC{}, false
		}
		i.Value = v
		return i, true

	case domain.IOCURL:
		u, err := url.Parse(strings.TrimSpace(i.Value))
		if err != nil || u.Scheme == "" || u.Host == "" {
			return domain.IOC{}, false
		}
		u.Scheme = strings.ToLower(u.Scheme)
		u.Host = strings.ToLower(u.Host)
		// Strip default ports.
		if (u.Scheme == "http" && strings.HasSuffix(u.Host, ":80")) || (u.Scheme == "https" && strings.HasSuffix(u.Host, ":443")) {
			u.Host = strings.Split(u.Host, ":")[0]
		}
		i.Value = u.String()
		return i, true

	case domain.IOCHash:
		v := strings.ToLower(spaceRe.ReplaceAllString(i.Value, ""))
		if !hexHashRe.MatchString(v) {
			return domain.IOC{}, false
		}
		i.Value = v
		return i, true
	default:
		return domain.IOC{}, false
	}
}

func NormalizeCampaign(c domain.Campaign) domain.Campaign {
	c.ID = strings.TrimSpace(c.ID)
	c.Name = strings.TrimSpace(c.Name)
	c.Summary = strings.TrimSpace(c.Summary)
	c.Source = strings.TrimSpace(c.Source)
	for i := range c.Actors {
		c.Actors[i] = strings.TrimSpace(c.Actors[i])
	}
	return c
}

func NormalizeCluster(cl domain.Cluster) domain.Cluster {
	cl.ID = strings.TrimSpace(cl.ID)
	cl.Name = strings.TrimSpace(cl.Name)
	cl.Description = strings.TrimSpace(cl.Description)
	cl.Source = strings.TrimSpace(cl.Source)
	return cl
}

