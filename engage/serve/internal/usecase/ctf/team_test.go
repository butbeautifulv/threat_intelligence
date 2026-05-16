package ctf

import "testing"

func TestTeamStrategy_assignsBySkill(t *testing.T) {
	c := NewCoordinator()
	challenges := []Challenge{
		{Name: "web1", Category: "web", Points: 200},
		{Name: "pwn1", Category: "pwn", Points: 300},
	}
	skills := map[string][]string{
		"alice": {"web"},
		"bob":   {"pwn"},
	}
	strategy := c.TeamStrategy(challenges, skills)
	if len(strategy.Assignments) != 2 {
		t.Fatalf("assignments %d", len(strategy.Assignments))
	}
}
