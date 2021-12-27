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
		debug     bool    // is debug enabled?
		// progress bool  // is progress enabled?
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
			// progress = stats.Progress
		}

		// Initialize the state table
		m = len(clauses)
		stateSize = 2*n + 2 + 3*m
		state = make([]State, stateSize)
		start = make([]int, m)
		size = make([]int, m)

		// index into state
		ptr := 2*n + 2

		// Iterate over clauses, last to first
		for i := range clauses {
			j := m - 1 - i // index into clauses

			start[i] = ptr
			size[i] = len(clauses[j])

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
				state[ptr].L = l
				state[ptr].C = j + 1
				state[l].C += 1

				// initialize the double linked list
				if state[l].F == 0 {
					state[l].F = ptr
					state[l].B = ptr
				}

				// insert into the beginning of the double linked list
				f, b := state[l].F, l
				state[ptr].F = f
				state[ptr].B = b
				state[b].F = ptr
				state[f].B = ptr

				// advance to the next position in the table
				ptr += 1
			}
		}

		if debug {
			dump()
		}
	}

	// A1 [Initialize.]
	if stats != nil && debug {
		log.Printf("A1. Initialize")
	}

	initialize()

	// 	var (
	// 		i int
	// 		j int
	// 		p int
	// 	)

	// 	if progress {
	// 		showProgress()
	// 	}

	// C2:
	// 	// C2. [Enter level l.]
	// 	if debug {
	// 		log.Printf("C2. Enter level %d, x[0:l]=%v\n", level, state[0:level])
	// 	}

	// 	if stats != nil {
	// 		stats.Levels[level]++
	// 		stats.Nodes++

	// 		if progress {
	// 			if level > stats.MaxLevel {
	// 				stats.MaxLevel = level
	// 			}
	// 			if stats.Nodes >= stats.Theta {
	// 				showProgress()
	// 				stats.Theta += stats.Delta
	// 			}
	// 		}
	// 	}

	// 	if rlink[0] == 0 {
	// 		// visit the solution
	// 		if debug {
	// 			log.Println("C2. Visit the solution")
	// 		}
	// 		if stats != nil {
	// 			stats.Solutions++
	// 		}
	// 		resume := lvisit()
	// 		if !resume {
	// 			if debug {
	// 				log.Println("C2. Halting the search")
	// 			}
	// 			if progress {
	// 				showProgress()
	// 			}
	// 			return nil
	// 		}
	// 		goto C8
	// 	}

	// 	// C3. [Choose i.]
	// 	if xccOptions.Exercise83 && level == 0 {
	// 		if debug && stats.Verbosity > 1 {
	// 			log.Print("Exercise 83: always choose i=1 at level=0")
	// 		}
	// 		i = 1
	// 	} else {
	// 		i = next_item()
	// 	}

	// 	if debug {
	// 		log.Printf("C3. Choose i=%d (%s)\n", i, name[i])
	// 	}

	// 	// C4. [Cover i.]
	// 	if debug {
	// 		log.Printf("C4. Cover i=%d (%s)\n", i, name[i])
	// 	}
	// 	cover(i)
	// 	state[level] = dlink[i]

	// C5:
	// 	// C5. [Try x_l.]
	// 	if debug {
	// 		log.Printf("C5. Try l=%d, x[0:l+1]=%v\n", level, state[0:level+1])
	// 	}
	// 	if state[level] == i {
	// 		goto C7
	// 	}
	// 	// Commit each of the items in this option
	// 	p = state[level] + 1
	// 	for p != state[level] {
	// 		j := top[p]
	// 		if j <= 0 {
	// 			// spacer, go back to the first option
	// 			p = ulink[p]
	// 		} else {
	// 			commit(p, j)
	// 			p++
	// 		}
	// 	}
	// 	level++
	// 	goto C2

	// C6:
	// 	// C6. [Try again.]
	// 	if debug {
	// 		log.Printf("C6. Try again, l=%d\n", level)
	// 	}

	// 	if stats != nil {
	// 		stats.Nodes++
	// 	}

	// 	// Uncommit each of the items in this option
	// 	p = state[level] - 1
	// 	for p != state[level] {
	// 		j = top[p]
	// 		if j <= 0 {
	// 			p = dlink[p]
	// 		} else {
	// 			uncommit(p, j)
	// 			p--
	// 		}
	// 	}

	// 	// Exercise 7.2.2.1-83
	// 	// This code works as expected for Exercise 87.  However, I am unable to
	// 	// reconcile my understanding of this answer to Exercise 83 with the actual
	// 	// description of the exercise.
	// 	// TODO: reconcile this discrepency
	// 	if xccOptions.Exercise83 && level == 0 {

	// 		// x is the first primary item covered
	// 		x := state[0]

	// 		// Find the spacer at the right of this option
	// 		for ; top[x] > 0; x++ {
	// 		}

	// 		// j is the last item in the option
	// 		j = top[x-1]

	// 		if j > n1 && color[x-1] == 0 {
	// 			// j is a secondary item with no color
	// 			// permanently remove from further consideration
	// 			if debug && stats.Verbosity > 1 {
	// 				log.Printf("Exercise 83: covering j=%d\n", j)
	// 			}
	// 			cover(j)
	// 			if debug && stats.Verbosity > 2 {
	// 				dump()
	// 			}
	// 		}

	// 	}

	// 	i = top[state[level]]
	// 	state[level] = dlink[state[level]]
	// 	goto C5

	// C7:
	// 	// C7. [Backtrack.]
	// 	if debug {
	// 		log.Println("C7. Backtrack")
	// 	}
	// 	uncover(i)

	// C8:
	// 	// C8. [Leave level l.]
	// 	if debug {
	// 		log.Printf("C8. Leaving level %d\n", level)
	// 	}
	// 	if level == 0 {
	// 		if progress {
	// 			showProgress()
	// 		}
	// 		return nil
	// 	}
	// 	level--
	// 	goto C6

	return nil
}
