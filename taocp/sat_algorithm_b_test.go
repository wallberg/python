package taocp

import (
	"log"
	"testing"
)

func TestSatAlgorithmB(t *testing.T) {

	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	cases := []struct {
		n       int        // number of strictly distinct literals
		sat     bool       // is satisfiable
		clauses SATClauses // clauses to satisfy
	}{
		{3, true, SATClauses{{1, -2}, {2, 3}, {-1, -3}, {-1, -2, 3}}},
		{3, false, SATClauses{{1, -2}, {2, 3}, {-1, -3}, {-1, -2, 3}, {1, 2, -3}}},
		{4, true, ClausesRPrime},
		// {4, false, ClausesR},
	}

	for _, c := range cases {

		stats := SATStats{
			Debug: true,
			// Progress: true,
		}
		options := SATOptions{}

		got, _ := SatAlgorithmB(c.n, c.clauses, &stats, &options)

		if got != c.sat {
			t.Errorf("expected satisfiable=%t for clauses %v; got %t", c.sat, c.clauses, got)
		}
	}
}

func TestSatAlgorithmBFromFile(t *testing.T) {

	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	cases := []struct {
		filename     string // file name of the SAT data file
		numVariables int    // number of variable
		numClauses   int    // number of clauses to satisfy
		sat          bool   // is satisfiable
	}{
		// {"testdata/SATExamples/A1.sat", 2043, 24772, false},
		// {"testdata/SATExamples/A2.sat", 2071, 25197, true},
		// {"testdata/SATExamples/K0.sat", 512, 5896, false},
		// {"testdata/SATExamples/P2.sat", 144, 530, false},
		// {"testdata/SATExamples/P3.sat", 144, 529, true},
		// {"testdata/SATExamples/P4.sat", 400, 2509, true},
		// {"testdata/SATExamples/L2.sat", 273, 1020, false},
		// {"testdata/SATExamples/L5.sat", 1472, 102922, true},
		{"testdata/SATExamples/L1.sat", 130, 2437, false},
		{"testdata/SATExamples/X2.sat", 129, 354, false},
	}

	for _, c := range cases {

		clauses, variables, err := ReadSAT(c.filename)

		// log.Printf("map=%v", literals)
		// for _, clause := range clauses {
		// 	log.Printf("%v", clause)
		// }

		if err != nil {
			t.Errorf("expected to read file %s; got error %v", c.filename, err)
			continue
		}
		if len(variables) != c.numVariables {
			t.Errorf("expected %d variables; got %d", c.numVariables, len(variables))
			continue
		}
		if len(clauses) != c.numClauses {
			t.Errorf("expected %d clauses; got %d", c.numClauses, len(clauses))
			continue
		}

		stats := SATStats{
			// Debug:    true,
			// Progress: true,
			// Delta:    100000000,
		}
		options := SATOptions{}

		got, _ := SatAlgorithmB(len(variables), clauses, &stats, &options)

		if got != c.sat {
			t.Errorf("expected satisfiable=%t for filename %s; got %t", c.sat, c.filename, got)
		}

	}
}
