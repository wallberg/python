package taocp

import (
	"fmt"
	"testing"
)

var ClausesR = SATClauses{
	{1, 2, -3},
	{2, 3, -4},
	{3, 4, 1},
	{4, -1, 2},
	{-1, -2, 3},
	{-2, -3, 4},
	{-3, -4, -1},
	{-4, 1, -2},
}

var ClausesRPrime = ClausesR[0:7]

func TestSATAlgorithmA(t *testing.T) {

	cases := []struct {
		n       int        // number of strictly distinct literals
		sat     bool       // is satisfiable
		clauses SATClauses // clauses to satisfy
	}{
		//{3, true, SATClauses{{1, -2}, {2, 3}, {-1, -3}, {-1, -2, 3}}},
		// {3, false, SATClauses{{1, -2}, {2, 3}, {-1, -3}, {-1, -2, 3}, {1, 2, -3}}},
		{4, true, ClausesRPrime},
	}

	for _, c := range cases {
		// if set, ok := sets.PieceSets[c.name]; !ok {
		// 	t.Errorf("Did not find set name='%s'", c.name)
		// } else {
		// 	if len(set) != c.count {
		// 		t.Errorf("Set '%s' has %d shapes; want %d",
		// 			c.name, len(set), c.count)
		// 	}
		// }

		stats := SATStats{
			Debug:    true,
			Progress: true,
		}
		options := SATOptions{}

		stats.Debug = true

		SATAlgorithmA(c.n, c.clauses, &stats, &options,
			func(solution [][]string) bool {
				fmt.Print(solution)
				return true
			})
	}
}
