package tinormalize

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/butbeautifulv/threat_intelligence/pipeline/internal/normalize/tidomain"
)

var (
	spaceRe   = regexp.MustCompile(`\s+`)
	hexHashRe = regexp.MustCompile(`^[a-f0-9]{32}$|^[a-f0-9]{40}$|^[a-f0-9]{64}$`)
)

func CanonicalID(i tidomain.IOC) string {
	h := sha256.Sum256([]byte(string(i.Type) + ":" + i.Value))
	return hex.EncodeToString(h[:])
}

func NormalizeIOC(i tidomain.IOC) (tidomain.IOC, bool) {
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
	case tidomain.IOCIP:
		ip := net.ParseIP(i.Value)
		if ip == nil {
			return tidomain.IOC{}, false
		}
		i.Value = ip.String()
		return i, true
	case tidomain.IOCDomain:
		v := strings.ToLower(strings.TrimSuffix(i.Value, "."))
		v = spaceRe.ReplaceAllString(v, "")
		if v == "" || strings.ContainsAny(v, "/:@") {
			return tidomain.IOC{}, false
		}
		i.Value = v
		return i, true
	case tidomain.IOCURL:
		u, err := url.Parse(strings.TrimSpace(i.Value))
		if err != nil || u.Scheme == "" || u.Host == "" {
			return tidomain.IOC{}, false
		}
		u.Scheme = strings.ToLower(u.Scheme)
		u.Host = strings.ToLower(u.Host)
		if (u.Scheme == "http" && strings.HasSuffix(u.Host, ":80")) || (u.Scheme == "https" && strings.HasSuffix(u.Host, ":443")) {
			u.Host = strings.Split(u.Host, ":")[0]
		}
		i.Value = u.String()
		return i, true
	case tidomain.IOCHash:
		v := strings.ToLower(spaceRe.ReplaceAllString(i.Value, ""))
		if !hexHashRe.MatchString(v) {
			return tidomain.IOC{}, false
		}
		i.Value = v
		return i, true
	default:
		return tidomain.IOC{}, false
	}
}

func NormalizeCampaign(c tidomain.Campaign) tidomain.Campaign {
	c.ID = strings.TrimSpace(c.ID)
	c.Name = strings.TrimSpace(c.Name)
	c.Summary = strings.TrimSpace(c.Summary)
	c.Source = strings.TrimSpace(c.Source)
	for i := range c.Actors {
		c.Actors[i] = strings.TrimSpace(c.Actors[i])
	}
	return c
}

func NormalizeCluster(cl tidomain.Cluster) tidomain.Cluster {
	cl.ID = strings.TrimSpace(cl.ID)
	cl.Name = strings.TrimSpace(cl.Name)
	cl.Description = strings.TrimSpace(cl.Description)
	cl.Source = strings.TrimSpace(cl.Source)
	return cl
}

func ActorStableID(name string) string {
	n := strings.TrimSpace(strings.ToLower(name))
	h := sha256.Sum256([]byte("actor:" + n))
	return hex.EncodeToString(h[:])
}

func ReportStableID(link string) string {
	h := sha256.Sum256([]byte("report:" + strings.TrimSpace(link)))
	return hex.EncodeToString(h[:])
}
