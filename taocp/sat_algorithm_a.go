package taocp

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
)

// SATAlgorithmA implements Algorithm A (7.2.2.2), satisfiability by backtracking.
// The task is to determine if the clause set is satisfiable, optionally
// return one or more satisfying assignments of the clauses.
//
// Arguments:
// n       -- number of strictly distinct literals
// clauses -- list of clauses to satisfy
// stats   -- SAT processing statistics
// options -- runtime options
// visit   -- function called with satisfying assignments; should return
//            true to request another assignment, false to halt
//
func SATAlgorithmA(n int, clauses SATClauses,
	stats *SATStats, options *SATOptions,
	visit func(solution [][]string) bool) error {

	// State represents a single cell in the state table
	type State struct {
		L int // literal
		F int // double linked list forward pointer to other cells with literal l
		B int // double linked list backward pointer to other cells with literal l
		C int // count of active clauses
	}

	var (
		m         int     // total number of clauses
		stateSize int     // total size of the state table
		state     []State // search state
		start     []int   // start of each clause in the table
		size      []int   // table of clause lengths
		a         int     // number of active clauses
		d         int     // depth-plus-one of the implicit search tree
		l         int     // literal
		p         int     // index into the state table
		i, j      int     // misc index values
		moves     []int   // store current progress
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

		// State, F
		b.WriteString("F(p) = ")
		for p := range state {
			if state[p].F == 0 {
				b.WriteString("  -")
			} else {
				b.WriteString(fmt.Sprintf(" %2d", state[p].F))
			}
		}
		b.WriteString("\n")

		// State, B
		b.WriteString("B(p) = ")
		for p := range state {
			if state[p].B == 0 {
				b.WriteString("  -")
			} else {
				b.WriteString(fmt.Sprintf(" %2d", state[p].B))
			}
		}
		b.WriteString("\n")

		// State, C
		b.WriteString("C(p) = ")
		for p := range state {
			if state[p].C == 0 {
				b.WriteString("  -")
			} else {
				b.WriteString(fmt.Sprintf(" %2d", state[p].C))
			}
		}
		b.WriteString("\n\n")

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

		// SIZE
		b.WriteString(" SIZE(i) = ")
		for _, val := range size {
			b.WriteString(fmt.Sprintf(" %2d", val))
		}
		b.WriteString("\n")

		log.Print(b.String())
	}

	// showProgress
	showProgress := func() {
		var b strings.Builder
		b.WriteString(fmt.Sprintf("d=%d, a=%d, moves=%v\n", d, a, moves))

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
		}

		// Initialize the state table
		m = len(clauses)
		stateSize = 2*n + 2 + 3*m
		state = make([]State, stateSize)
		start = make([]int, m+1)
		size = make([]int, m+1)
		moves = make([]int, n+1)

		// index into state
		p := 2*n + 2

		// Iterate over clauses, last to first
		for i := range clauses {
			j := m - 1 - i // index into clauses

			start[i+1] = p
			size[i+1] = len(clauses[j])

			// Sort literals of the clause in descending order
			clause := make(SATClause, len(clauses[j]))
			copy(clause, clauses[j])
			sort.SliceStable(clause, func(i, j int) bool {
				// Sort by the absolute value of the literal, descending
				return math.Abs(float64(clause[j])) < math.Abs(float64(clause[i]))
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
				state[p].C = j + 1
				state[l].C += 1

				// initialize the double linked list
				if state[l].F == 0 {
					state[l].F = p
					state[l].B = p
				}

				// insert into the beginning of the double linked list
				f, b := state[l].F, l
				state[p].F = f
				state[p].B = b
				state[b].F = p
				state[f].B = p

				// advance to the next position in the table
				p += 1
			}
		}

		if stats.Debug {
			dump()
		}
	}

	//
	// A1 [Initialize.]
	//
	if stats != nil && stats.Debug {
		log.Printf("A1. Initialize")
	}

	initialize()

	a = m
	d = 1

	if stats.Progress {
		showProgress()
	}

A2:
	//
	// A2. [Choose.]
	//
	if stats.Debug {
		log.Printf("A2. Choose.")
	}

	// if stats != nil {
	// 	stats.Levels[d-1]++
	// 	stats.Nodes++

	// 	if stats.Progress {
	// 		if level > stats.MaxLevel {
	// 			stats.MaxLevel = level
	// 		}
	// 		if stats.Nodes >= stats.Theta {
	// 			showProgress()
	// 			stats.Theta += stats.Delta
	// 		}
	// 	}
	// }

	l = 2 * d
	if state[l].C <= state[l+1].C {
		l += 1
	}

	moves[d] = l & 1
	if l^1 == 0 {
		moves[d] += 4
	}

	showProgress()

	if state[l].C == a {
		// // visit the solution
		// if stats.Debug {
		// 	log.Println("C2. Visit the solution")
		// }
		// if stats != nil {
		// 	stats.Solutions++
		// }
		// resume := lvisit()
		// if !resume {
		// 	if stats.Debug {
		// 		log.Println("C2. Halting the search")
		// 	}
		// 	if stats.Progress {
		// 		showProgress()
		// 	}
		// 	return nil
		// }

		return nil
	}

A3:
	//
	// A3. [Remove ^l.]
	//
	if stats.Debug {
		log.Printf("A3. Remove ^l.")
	}

	// Delete ^l from all active clauses; that is, ignore ^l because
	// we are making l true

	// Start at the very beginning
	p = state[l^1].F

	// Iterate over the clauses containing ^l
	for p >= 2*n+2 {
		j = state[p].C
		i = size[j]
		if i > 1 {
			// Remove ^l from this clause
			size[j] -= 1

			// Advance to next clause
			p = state[p].F

		} else if i == 1 {
			// ^l is the last literal and would make the clause empty
			// undo what we've just done and go to A5

			// Reverse direction
			p = state[p].B

			// Iterate back through the clauses
			for p >= 2*n+2 {
				// Add ^l back to the clause
				j = state[p].C
				size[j] += 1

				// Advance to the next clause
				p = state[p].B
			}

			goto A5

		} else {
			log.Fatal("Should not be reachable")
		}
	}

	// A4. [Deactivate l's clauses.]
	if stats.Debug {
		log.Printf("A4. [Deactivate l's clauses.]")
	}

	// Suppress all clauses tht contain l

	a -= state[l].C
	d += 1

	goto A2

A5:
	// A5 [Try again.]
	if stats.Debug {
		log.Printf("A5 [Try again.]")
	}

	if moves[d] < 2 {
		moves[d] = 3 - moves[d]
		l = 2*d + (moves[d] & 1)
		goto A3
	}

	// A6 [Backtrack.]
	if stats.Debug {
		log.Printf("A6 [Backtrack.]")
	}

	if d == 1 {
		// unsatisfiable
		return nil
	}

	d -= 1
	l = 2*d + (moves[d] & 1)

	// A7 [Reactivate l's clauses.]
	if stats.Debug {
		log.Printf("A7 [Reactivate l's clauses.]")
	}

	a += state[l].C

	// Unsuppress all clauses that contain l.

	// A8 [Unremove ^l.]
	if stats.Debug {
		log.Printf("A8 [Unremove ^l.]")
	}

	// Reinstate ^l in all the active clauses that contain it.

	goto A5

}
