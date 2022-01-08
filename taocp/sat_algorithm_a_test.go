package taocp

import (
	"log"
	"testing"
)

func TestSatAlgorithmA(t *testing.T) {

	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	cases := []struct {
		n       int        // number of strictly distinct literals
		sat     bool       // is satisfiable
		clauses SatClauses // clauses to satisfy
	}{
		{3, true, SatClauses{{1, -2}, {2, 3}, {-1, -3}, {-1, -2, 3}}},
		{3, false, SatClauses{{1, -2}, {2, 3}, {-1, -3}, {-1, -2, 3}, {1, 2, -3}}},
		{4, true, ClausesRPrime},
		{4, false, ClausesR},
		{9, false, ClausesWaerden339},
	}

	for _, c := range cases {

		stats := SatStats{
			// Debug: true,
			// Progress: true,
		}
		options := SatOptions{}

		got, _ := SatAlgorithmA(c.n, c.clauses, &stats, &options)

		if got != c.sat {
			t.Errorf("expected satisfiable=%t for clauses %v; got %t", c.sat, c.clauses, got)
		}
	}
}

func TestSatAlgorithmAFromFile(t *testing.T) {

	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	cases := []struct {
		filename     string // file name of the SAT data file
		numVariables int    // number of variable
		numClauses   int    // number of clauses to satisfy
		sat          bool   // is satisfiable
	}{
		{"testdata/SATExamples/L1.sat", 130, 2437, false},
		{"testdata/SATExamples/X2.sat", 129, 354, false},
	}

	for _, c := range cases {

		t.Run(c.filename, func(t *testing.T) {
			clauses, variables, err := SatRead(c.filename)

			// log.Printf("map=%v", literals)
			// for _, clause := range clauses {
			// 	log.Printf("%v", clause)
			// }

			if err != nil {
				t.Errorf("expected to read file %s; got error %v", c.filename, err)
				return
			}
			if len(variables) != c.numVariables {
				t.Errorf("expected %d variables; got %d", c.numVariables, len(variables))
				return
			}
			if len(clauses) != c.numClauses {
				t.Errorf("expected %d clauses; got %d", c.numClauses, len(clauses))
				return
			}

			stats := SatStats{
				// Debug:    true,
				// Progress: true,
				// Delta:    100000000,
			}
			options := SatOptions{}

			got, _ := SatAlgorithmA(len(variables), clauses, &stats, &options)

			if got != c.sat {
				t.Errorf("expected satisfiable=%t for filename %s; got %t", c.sat, c.filename, got)
			}
		})
	}
}
