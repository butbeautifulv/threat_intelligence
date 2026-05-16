package intelligence

import (
	"net/http"
	"strings"
)

// headerSignatures maps header/value needles to technologies (HexStrike subset).
var headerSignatures = map[string][]Technology{
	"apache":    {TechApache},
	"nginx":     {TechNginx},
	"microsoft": {TechIIS},
	"iis":       {TechIIS},
	"php":       {TechPHP},
	"express":   {TechNodeJS},
	"django":    {TechPython},
	"flask":     {TechPython},
	"werkzeug":  {TechPython},
	"tomcat":    {TechJava},
	"jboss":     {TechJava},
	"weblogic":  {TechJava},
	"asp.net":   {TechDotNet},
}

// contentSignatures maps body needles to technologies.
var contentSignatures = map[string][]Technology{
	"wp-content":        {TechWordPress, TechPHP},
	"wp-includes":       {TechWordPress},
	"wordpress":         {TechWordPress},
	"drupal":            {TechDrupal},
	"/sites/default":    {TechDrupal},
	"joomla":            {TechJoomla},
	"/administrator":    {TechJoomla},
	"react":             {TechReact},
	"__react_devtools":  {TechReact},
	"angular":           {TechAngular},
	"ng-version":        {TechAngular},
	"vue":               {TechVue},
	"__vue__":           {TechVue},
}

// MatchHeaderSignatures detects technologies from HTTP headers.
func MatchHeaderSignatures(hdr http.Header) []Technology {
	if hdr == nil {
		return nil
	}
	seen := map[Technology]struct{}{}
	add := func(ts ...Technology) {
		for _, t := range ts {
			seen[t] = struct{}{}
		}
	}
	check := func(val string) {
		low := strings.ToLower(val)
		for needle, techs := range headerSignatures {
			if strings.Contains(low, needle) {
				add(techs...)
			}
		}
	}
	check(hdr.Get("Server"))
	check(hdr.Get("X-Powered-By"))
	check(hdr.Get("X-AspNet-Version"))
	return technologiesFromSet(seen)
}

// MatchContentSignatures detects technologies from HTML/body sample.
func MatchContentSignatures(body string) []Technology {
	if body == "" {
		return nil
	}
	low := strings.ToLower(body)
	seen := map[Technology]struct{}{}
	for needle, techs := range contentSignatures {
		if strings.Contains(low, needle) {
			for _, t := range techs {
				seen[t] = struct{}{}
			}
		}
	}
	return technologiesFromSet(seen)
}

func technologiesFromSet(seen map[Technology]struct{}) []Technology {
	if len(seen) == 0 {
		return nil
	}
	out := make([]Technology, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	return out
}
