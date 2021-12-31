package taocp

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

// ReadSAT reads a SAT file in Knuth format and returns
// a list of clauses along with the mapping of variables
// (numeric to string name)
func ReadSAT(filename string) (SATClauses, map[int]string, error) {

	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening %s for reading: %v", filename, err)
	}

	var clauses SATClauses
	variable2name := make(map[int]string)
	name2variable := make(map[string]int)
	nextVariable := 1 // next literal to use

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Check if a comment line
		if !strings.HasPrefix(line, "~ ") {
			// Create a new clause
			var clause SATClause

			// Iterate over the literals of the clause
			for _, name := range strings.Fields(line) {

				// Determine if the literal is negated
				sign := 1
				if strings.HasPrefix(name, "~") {
					name = name[1:]
					sign = -1
				}

				// Determine the variable number for this name
				var found bool
				var variable int
				if variable, found = name2variable[name]; !found {
					variable = nextVariable
					name2variable[name] = variable
					variable2name[variable] = name
					nextVariable += 1
				}

				literal := sign * variable

				clause = append(clause, literal)
			}
			clauses = append(clauses, clause)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error scanning %s: %v", filename, err)
	}

	return clauses, variable2name, nil
}
