package taocp

import (
	"fmt"
)

// Explore Satisfiability from The Art of Computer Programming, Volume 4,
// Fascicle 6, Satisfiability, 2015
//
// §7.2.2.2 Satisfiability (SAT)

// SATStats is a struct for tracking SAT statistics and reporting
// runtime progress
type SATStats struct {
	// Input parameters
	Progress     bool // Display runtime progress
	Debug        bool // Enable debug logging
	Verbosity    int  // Debug verbosity level (0 or 1)
	Delta        int  // Display progress every Delta number of Nodes
	SuppressDump bool // Don't display the dump()

	// Statistics collectors
	MaxLevel  int   // Maximum level reached
	Theta     int   // Display progress at next Theta number of Nodes
	Levels    []int // Count of times each level is entered
	Nodes     int   // Count of nodes processed
	Solutions int   // Count of solutions returned
}

// SATOptions provides SAT runtime options
type SATOptions struct {
}

// String returns a String representation of type SATStats struct
func (s SATStats) String() string {
	// Find first non-zero level count
	i := len(s.Levels)
	for s.Levels[i-1] == 0 && i > 1 {
		i--
	}

	return fmt.Sprintf("nodes=%d, solutions=%d, levels=%v", s.Nodes,
		s.Solutions, s.Levels[:i])
}

// SATClause represents a single clause
type SATClause []int

// SATClauses represents a list of clauses
type SATClauses []SATClause
