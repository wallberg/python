package taocp

import (
	"reflect"
	"testing"
)

var ClausesR = SatClauses{
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

var ClausesWaerden339 = SatWaerdan(3, 3, 9)

func TestReadSAT(t *testing.T) {

	cases := []struct {
		filename     string // file name of the SAT data file
		numVariables int    // number of variable
		numClauses   int    // number of clauses to satisfy
	}{
		{"testdata/SATExamples/test.sat", 19, 45},
		{"testdata/SATExamples/A1.sat", 2043, 24772},
		{"testdata/SATExamples/A2.sat", 2071, 25197},
	}

	for _, c := range cases {

		clauses, variables, err := SatRead(c.filename)

		// log.Printf("map=%v", literals)
		// for _, clause := range clauses {
		// 	log.Printf("%v", clause)
		// }

		if err != nil {
			t.Errorf("expected to read file %s; got error %v", c.filename, err)
		} else {
			if len(variables) != c.numVariables {
				t.Errorf("expected %d variables; got %d", c.numVariables, len(variables))
			}
			if len(clauses) != c.numClauses {
				t.Errorf("expected %d clauses; got %d", c.numClauses, len(clauses))
			}
		}
	}
}

func TestSatWaerden(t *testing.T) {
	cases := []struct {
		j, k, n int
		clauses SatClauses
	}{
		{3, 3, 9, SatClauses{
			{1, 2, 3}, {2, 3, 4}, {3, 4, 5}, {4, 5, 6}, {5, 6, 7}, {6, 7, 8}, {7, 8, 9},
			{1, 3, 5}, {2, 4, 6}, {3, 5, 7}, {4, 6, 8}, {5, 7, 9},
			{1, 4, 7}, {2, 5, 8}, {3, 6, 9},
			{1, 5, 9},
			{-1, -2, -3}, {-2, -3, -4}, {-3, -4, -5}, {-4, -5, -6}, {-5, -6, -7}, {-6, -7, -8}, {-7, -8, -9},
			{-1, -3, -5}, {-2, -4, -6}, {-3, -5, -7}, {-4, -6, -8}, {-5, -7, -9},
			{-1, -4, -7}, {-2, -5, -8}, {-3, -6, -9},
			{-1, -5, -9}}},
	}

	for _, c := range cases {
		got := SatWaerdan(c.j, c.k, c.n)

		if !reflect.DeepEqual(got, c.clauses) {
			t.Errorf("expected for waerden(%d,%d,%d) clauses %v; got %v", c.j, c.k, c.n, c.clauses, got)
		}
	}
}
