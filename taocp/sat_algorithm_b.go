package taocp

import (
	"fmt"
	"log"
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
		i, j, k   int     // indices
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

		// l
		b.WriteString("       l = ")
		for l := range watch {
			b.WriteString(fmt.Sprintf(" %2d", l))
		}
		b.WriteString("\n")

		// WATCH
		b.WriteString("WATCH(l) = ")
		for _, val := range watch {
			b.WriteString(fmt.Sprintf(" %2d", val))
		}
		b.WriteString("\n")

		// j
		b.WriteString("       j = ")
		for j := range start {
			b.WriteString(fmt.Sprintf(" %2d", j))
		}
		b.WriteString("\n")

		// START
		b.WriteString("START(j) = ")
		for _, val := range start {
			b.WriteString(fmt.Sprintf(" %2d", val))
		}
		b.WriteString("\n")

		// LINK
		b.WriteString(" LINK(j) = ")
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
		watch = make([]int, 2*n+2)
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
			clause := clauses[j-1]
			start[j] = p - len(clause) + 1

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

				// Check if this is the last literal in the clause
				if p == start[j] {
					// Insert this literal into the watch list of clauses.
					// watch is the head of the list, with link containing
					// the next pointers. The last clause in the list has
					// value of 0.
					if watch[l] == 0 {
						// Insert the first clause into the list
						watch[l] = j
					} else {
						// Insert this clause to the end of the list
						jp := watch[l]
						for link[jp] != 0 {
							jp = link[jp]
						}
						link[jp] = j
					}
				}

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

	//
	// B2. [Rejoice or choose.]
	//
B2:
	if d > n {
		// visit the solution
		if debug {
			log.Println("B2. [Rejoice.]")
		}
		if stats != nil {
			stats.Solutions++
		}

		return true, lvisit()
	}

	if watch[2*d] == 0 || watch[2*d+1] != 0 {
		moves[d] = 1
	} else {
		moves[d] = 0
	}

	l = 2*d + moves[d]

	if debug {
		log.Printf("B2. [Choose.] d=%d, l=%d, moves=%v", d, l, moves[1:d+1])
	}

	if stats != nil {
		stats.Levels[d-1]++
		stats.Nodes++

		if progress {
			if d > stats.MaxLevel {
				stats.MaxLevel = d
			}
			if stats.Nodes >= stats.Theta {
				showProgress()
				stats.Theta += stats.Delta
			}
		}
	}

	//
	// B3. [Remove ^l if possible.]
	//
B3:
	if debug {
		log.Printf("B3. [Remove ^l if possible.] ^l=%d.", l^1)
	}

	// For all j such that ^l is watched in clause j, watch another
	// literal of clause j. But go to B5 if that can't be done.

	j = watch[l^1]
	if debug {
		log.Printf("B3.   j=watch[%d]=%d", j, l^1)
	}

	// While j <> 0, a literal other than ^l should be watch in clause j
	for j != 0 {
		i = start[j]
		ip := start[j-1]
		jp := link[j]
		k = i + 1

		if debug {
			log.Printf("B3.   j=%d, i=%d, ip=%d, jp=%d", j, i, ip, jp)
		}

		for k < ip {
			lp := state[k].L

			if debug {
				log.Printf("B3.   k=%d, lp=%d", k, lp)
			}

			// check if lp isn't false
			if (lp>>1) > d || (lp+moves[lp>>1])%2 == 0 {
				state[i].L = lp
				state[k].L = l ^ 1
				link[j] = watch[lp]
				watch[lp] = j
				j = jp
				break
			}

			k += 1
		}

		if k == ip {
			// we can't stop watching ^l
			watch[l^1] = j

			goto B5
		}

	}

	//
	// B4. [Advance.]
	//

	watch[l^1] = 0
	d += 1

	if debug {
		log.Printf("B4. [Advance.] d=%d", d)
	}

	goto B2

B5:
	//
	// B5. [Try again.]
	//
	if debug {
		log.Printf("B5. [Try again.]")
	}

	if moves[d] < 2 {
		moves[d] = 3 - moves[d]
		l = 2*d + (moves[d] & 1)

		if debug {
			log.Printf("B5.   d=%d, l=%d, moves=%v", d, l, moves[1:d+1])
		}

		if stats != nil {
			stats.Nodes++
		}

		goto B3
	}

	//
	// B6. [Backtrack.]
	//
	if debug {
		log.Printf("B6. [Backtrack.]")
	}

	if d == 1 {
		// unsatisfiable
		return false, nil
	}

	// Decrement the depth
	d -= 1

	if debug {
		log.Printf("B6. d=%d, l=%d, moves=%v", d, l, moves[1:d+1])
	}

	goto B5
}
