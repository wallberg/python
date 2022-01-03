package taocp

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
)

// SatAlgorithmB implements Algorithm B (7.2.2.2), satisfiability by watching.
// The task is to determine if the clause set is satisfiable, and if it is return
// one satisfying assignment of the clauses.
//
// Arguments:
// n       -- number of strictly distinct literals
// clauses -- list of clauses to satisfy
// stats   -- SAT processing statistics
// options -- runtime options
//
func SatAlgorithmB(n int, clauses SATClauses,
	stats *SATStats, options *SATOptions) (bool, []int) {

	// State represents a single cell in the state table
	type State struct {
		L int // literal
	}

	var (
		m         int     // total number of clauses
		stateSize int     // total size of the state table
		state     []State // search state
		start     []int   // start of each clause in the table
		watch     []int   // list of all clauses that currently watch l
		link      []int   // the number of another clause with the same watch literal
		d         int     // depth-plus-one of the implicit search tree
		l         int     // literal
		p         int     // index into the state table
		moves     []int   // store current progress
		debug     bool    // debugging is enabled
		progress  bool    // progress tracking is enabled
	)

	// dump
	dump := func() {

		var b strings.Builder
		b.WriteString("\n")

		// State, p
		b.WriteString("   p = ")
		for p := range state {
			b.WriteString(fmt.Sprintf(" %2d", p))
		}
		b.WriteString("\n")

		// State, L
		b.WriteString("L(p) = ")
		for p := range state {
			if state[p].L == 0 {
				b.WriteString("  -")
			} else {
				b.WriteString(fmt.Sprintf(" %2d", state[p].L))
			}
		}
		b.WriteString("\n")

		// i
		b.WriteString("       i = ")
		for i := range start {
			b.WriteString(fmt.Sprintf(" %2d", i))
		}
		b.WriteString("\n")

		// START
		b.WriteString("START(i) = ")
		for _, val := range start {
			b.WriteString(fmt.Sprintf(" %2d", val))
		}
		b.WriteString("\n")

		// WATCH
		b.WriteString("WATCH(i) = ")
		for _, val := range watch {
			b.WriteString(fmt.Sprintf(" %2d", val))
		}
		b.WriteString("\n")

		// LINK
		b.WriteString(" LINK(i) = ")
		for _, val := range link {
			b.WriteString(fmt.Sprintf(" %2d", val))
		}
		b.WriteString("\n")

		log.Print(b.String())
	}

	// showProgress
	showProgress := func() {
		var b strings.Builder
		b.WriteString(fmt.Sprintf("Nodes=%d, d=%d, l=%d, moves=%v\n", stats.Nodes, d, l, moves[1:d+1]))

		log.Print(b.String())
	}

	// initialize
	initialize := func() {

		if stats != nil {
			stats.Theta = stats.Delta
			stats.MaxLevel = -1
			if stats.Levels == nil {
				stats.Levels = make([]int, n)
			} else {
				for len(stats.Levels) < n {
					stats.Levels = append(stats.Levels, 0)
				}
			}
			debug = stats.Debug
			progress = stats.Progress
		}

		// Initialize the state table
		m = len(clauses)
		start = make([]int, m+1)
		link = make([]int, m+1)
		watch = make([]int, m+1)
		moves = make([]int, n+1)

		for _, clause := range clauses {
			stateSize += len(clause)
		}
		state = make([]State, stateSize)

		start[0] = stateSize

		// index into state
		p = stateSize - 1

		// Iterate over the clauses
		for j := 1; j <= len(clauses); j++ {
			clauseLen := len(clauses[j-1])
			start[j] = p - clauseLen + 1

			// Sort literals of the clause in ascending order
			clause := make(SATClause, clauseLen)
			copy(clause, clauses[j-1])
			sort.SliceStable(clause, func(i, j int) bool {
				// Sort by the absolute value of the literal, descending
				return math.Abs(float64(clause[i])) < math.Abs(float64(clause[j]))
			})

			// Iterate over literal values of the clauses
			for _, k := range clause {
				// compute literal l
				var l int
				if k >= 0 {
					l = 2 * k
				} else {
					l = -2*k + 1
				}

				// insert into the state table
				state[p].L = l

				// advance to the next position in the table
				p -= 1
			}
		}

		if debug {
			dump()
		}
	}

	// lvisit prepares the solution
	lvisit := func() []int {
		solution := make([]int, n)
		for i := 1; i < n+1; i++ {
			solution[i-1] = (moves[i] % 2) ^ 1
		}
		if debug {
			log.Printf("visit solution=%v", solution)
		}

		return solution
	}

	//
	// B1 [Initialize.]
	//

	initialize()

	if debug {
		log.Printf("B1. Initialize")
	}

	d = 1

	if debug {
		log.Printf("    d=%d, l=%d, moves=%v", d, l, moves[1:d+1])
	}

	if progress {
		showProgress()
	}

	return true, lvisit()

	// A2:
	// 	//
	// 	// A2. [Choose.]
	// 	//

	// 	// Choose l if it is contained in the most clauses, other ^l
	// 	l = 2 * d
	// 	if state[l].C <= state[l+1].C {
	// 		l += 1
	// 	}

	// 	moves[d] = l & 1
	// 	if state[l^1].C == 0 {
	// 		moves[d] += 4
	// 	}

	// 	if debug {
	// 		log.Printf("A2. [Choose.] d=%d, a=%d, l=%d, moves=%v", d, a, l, moves[1:d+1])
	// 	}

	// 	if stats != nil {
	// 		stats.Levels[d-1]++
	// 		stats.Nodes++

	// 		if progress {
	// 			if d > stats.MaxLevel {
	// 				stats.MaxLevel = d
	// 			}
	// 			if stats.Nodes >= stats.Theta {
	// 				showProgress()
	// 				stats.Theta += stats.Delta
	// 			}
	// 		}
	// 	}

	// 	if state[l].C == a {
	// 		// visit the solution
	// 		if debug {
	// 			log.Println("A2.   Visit the solution")
	// 		}
	// 		if stats != nil {
	// 			stats.Solutions++
	// 		}

	// 		return true, lvisit()
	// 	}

	// A3:
	// 	//
	// 	// A3. [Remove ^l.]
	// 	//
	// 	if debug {
	// 		log.Printf("A3. [Remove ^l.] ^l=%d.", l^1)
	// 	}

	// 	// Delete ^l from all active clauses; that is, ignore ^l because
	// 	// we are making l true

	// 	// Start at the first clause containing ^l
	// 	p = state[l^1].F

	// 	// Iterate over the clauses containing ^l
	// 	for p >= 2*n+2 {
	// 		j := state[p].C
	// 		i := size[j]

	// 		if i > 1 {
	// 			// Remove ^l from this clause
	// 			size[j] = i - 1

	// 			// Advance to next clause
	// 			p = state[p].F

	// 		} else if i == 1 {
	// 			// ^l is the last literal and would make the clause empty
	// 			// undo what we've just done and go to A5

	// 			if debug {
	// 				log.Printf("A3. Cancel, this would leave a clause empty; p=%d, j=%d, i=%d", p, j, i)
	// 			}

	// 			// Reverse direction
	// 			p = state[p].B

	// 			// Iterate back through the clauses
	// 			for p >= 2*n+2 {
	// 				// Add ^l back to the clause
	// 				j = state[p].C
	// 				size[j] += 1

	// 				// Advance to the next clause
	// 				p = state[p].B
	// 			}

	// 			goto A5

	// 		} else {
	// 			log.Fatal("A3. Should not be reachable")
	// 		}
	// 	}

	// 	//
	// 	// A4. [Deactivate l's clauses.]
	// 	//
	// 	if debug {
	// 		log.Printf("A4. [Deactivate l's clauses.] l=%d", l)
	// 	}

	// 	// Suppress all clauses that contain l

	// 	// Start at the first clause containing l
	// 	p = state[l].F

	// 	// Iterate over the clauses containing l
	// 	for p >= 2*n+2 {
	// 		j := state[p].C
	// 		i := start[j]

	// 		// Iterate over each literal and remove from the clause
	// 		for s := i; s < i+size[j]-1; s++ {
	// 			f, b := state[s].F, state[s].B
	// 			state[f].B = b
	// 			state[b].F = f
	// 			state[state[s].L].C -= 1
	// 			if state[state[s].L].C < 0 {
	// 				dump()
	// 				log.Fatal("A4. C(L(s)) should not be < 0")
	// 			}
	// 		}

	// 		p = state[p].F

	// 	}

	// 	// Update count of total active clauses
	// 	a -= state[l].C

	// 	// Increment the depth
	// 	d += 1

	// 	goto A2

	// A5:
	// 	//
	// 	// A5. [Try again.]
	// 	//
	// 	if debug {
	// 		log.Printf("A5. [Try again.]")
	// 	}

	// 	if moves[d] < 2 {
	// 		moves[d] = 3 - moves[d]
	// 		l = 2*d + (moves[d] & 1)

	// 		if debug {
	// 			log.Printf("A5. d=%d, a=%d, l=%d, moves=%v", d, a, l, moves[1:d+1])
	// 		}

	// 		if stats != nil {
	// 			stats.Nodes++
	// 		}

	// 		goto A3
	// 	}

	// 	//
	// 	// A6. [Backtrack.]
	// 	//
	// 	if debug {
	// 		log.Printf("A6. [Backtrack.]")
	// 	}

	// 	if d == 1 {
	// 		// unsatisfiable
	// 		return false, nil
	// 	}

	// 	// Decrement the depth
	// 	d -= 1

	// 	// TODO: what are we doing?
	// 	l = 2*d + (moves[d] & 1)

	// 	if debug {
	// 		log.Printf("A6. d=%d, a=%d, l=%d, moves=%v", d, a, l, moves[1:d+1])
	// 	}

	// 	//
	// 	// A7 [Reactivate l's clauses.]
	// 	//
	// 	if debug {
	// 		log.Printf("A7. [Reactivate l's clauses.]")
	// 	}

	// 	// Update count of total active clauses
	// 	a += state[l].C

	// 	// Unsuppress all clauses that contain l.

	// 	// Start at the last clause containing l
	// 	p = state[l].B

	// 	// Iterate over the clauses containing l
	// 	for p >= 2*n+2 {
	// 		j := state[p].C
	// 		i := start[j]

	// 		// Iterate over each literal and add back to the clause
	// 		for s := i; s < i+size[j]-1; s++ {
	// 			f, b := state[s].F, state[s].B
	// 			state[f].B = s
	// 			state[b].F = s
	// 			state[state[s].L].C += 1
	// 		}

	// 		// Advance to the next clause
	// 		p = state[p].B
	// 	}

	// 	if debug {
	// 		log.Printf("A7. d=%d, a=%d, l=%d, moves=%v", d, a, l, moves[1:d+1])
	// 	}

	// 	//
	// 	// A8. [Unremove ^l.]
	// 	//
	// 	if debug {
	// 		log.Printf("A8. [Unremove ^l.]")
	// 	}

	// 	// Reinstate ^l in all the active clauses that contain it.

	// 	// Start at the first clause containing ^l
	// 	p = state[l^1].F

	// 	// Iterate over the clauses containing l
	// 	for p >= 2*n+2 {
	// 		j := state[p].C
	// 		size[j] += 1

	// 		// Advance to the next clause
	// 		p = state[p].F
	// 	}

	// 	goto A5

}
