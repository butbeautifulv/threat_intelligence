package ctf

import (
	"sort"
	"strings"
)

// TeamStrategy assigns challenges to team members by skill.
type TeamStrategy struct {
	Assignments      []Assignment `json:"assignments"`
	Unassigned       []string     `json:"unassigned,omitempty"`
	EstimatedTotal   int          `json:"estimated_total_seconds"`
	RecommendedOrder []string     `json:"recommended_order"`
}

// Assignment maps a challenge to a team member.
type Assignment struct {
	Challenge  string `json:"challenge"`
	Category   string `json:"category"`
	Assignee   string `json:"assignee"`
	Points     int    `json:"points"`
	Priority   int    `json:"priority"`
}

// Coordinator builds team strategies for CTF events.
type Coordinator struct {
	Manager *Manager
}

func NewCoordinator() *Coordinator {
	return &Coordinator{Manager: NewManager()}
}

// TeamStrategy distributes challenges across team_skills map (member -> []categories).
func (c *Coordinator) TeamStrategy(challenges []Challenge, teamSkills map[string][]string) TeamStrategy {
	out := TeamStrategy{Assignments: []Assignment{}}
	if len(challenges) == 0 {
		return out
	}
	assigned := map[string]struct{}{}
	memberLoad := map[string]int{}

	// Sort challenges by points desc
	sorted := append([]Challenge(nil), challenges...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Points > sorted[j].Points
	})

	for _, ch := range sorted {
		bestMember := ""
		bestScore := -1
		for member, skills := range teamSkills {
			score := skillMatch(ch.Category, skills)
			load := memberLoad[member]
			adj := score - load*2
			if adj > bestScore {
				bestScore = adj
				bestMember = member
			}
		}
		if bestMember == "" {
			out.Unassigned = append(out.Unassigned, ch.Name)
			continue
		}
		assigned[ch.Name] = struct{}{}
		memberLoad[bestMember]++
		wf := c.Manager.CreateChallengeWorkflow(ch, nil)
		out.Assignments = append(out.Assignments, Assignment{
			Challenge: ch.Name,
			Category:  ch.Category,
			Assignee:  bestMember,
			Points:    ch.Points,
			Priority:  ch.Points / 10,
		})
		out.EstimatedTotal += wf.EstimatedTime
		out.RecommendedOrder = append(out.RecommendedOrder, ch.Name)
	}
	return out
}

func skillMatch(category string, skills []string) int {
	category = strings.ToLower(category)
	score := 0
	for _, s := range skills {
		s = strings.ToLower(strings.TrimSpace(s))
		if s == category || s == "all" {
			score += 10
		}
		if strings.Contains(category, s) || strings.Contains(s, category) {
			score += 5
		}
	}
	return score
}
