package taocp

import (
	"fmt"
	"log"
	"strings"
)

// SatAlgorithmD implements Algorithm D (7.2.2.2), satisfiability by cyclic DPLL.
// The task is to determine if the clause set is satisfiable, and if it is return
// one satisfying assignment of the clauses.
//
// Arguments:
// n       -- number of strictly distinct literals
// clauses -- list of clauses to satisfy
// stats   -- SAT processing statistics
// options -- runtime options
//
func SatAlgorithmD(n int, clauses SatClauses,
	stats *SatStats, options *SatOptions) (bool, []int) {

	// State represents a single cell in the state table
	type State struct {
		L int // literal
	}

	var (
		m           int     // total number of clauses
		stateSize   int     // total size of the state table
		state       []State // search state
		start       []int   // start of each clause in the table
		watch       []int   // list of all clauses that currently watch l
		link        []int   // the number of another clause with the same watch literal
		h           []int   // the literal being watched at depth d
		next        []int   // active ring : not-yet-set variables whose watch lists aren't empty
		head        int     // head pointer into the active ring
		tail        int     // tail pointer into the active ring
		d           int     // depth of the implicit search tree
		x           []int   // selected literal at depth d
		l, lp       int     // literal
		p, q        int     // index into the state table
		b           int     // branch on literal?
		f           int     // flag?
		i, j, jp, k int     // indices
		moves       []int   // store current progress
		debug       bool    // debugging is enabled
		progress    bool    // progress tracking is enabled
	)

	// activeRing returns a string with members of the active ring
	activeRing := func() string {
		var b strings.Builder
		b.WriteString("(")

		if tail != 0 {
			l := head
			for {
				if l != head {
					b.WriteString(" ")
				}
				b.WriteString(fmt.Sprintf("%d", l))
				l = next[l]
				if l == head {
					break
				}
			}
		}
		b.WriteString(")")

		return b.String()
	}

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

		// NEXT
		b.WriteString(" NEXT(v) = ")
		for _, val := range next {
			b.WriteString(fmt.Sprintf(" %2d", val))
		}
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("head=%d, tail=%d\n", head, tail))
		b.WriteString(fmt.Sprintf("active ring=%s\n", activeRing()))

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
		h = make([]int, n+1)
		next = make([]int, n+1)
		x = make([]int, n+1)

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
	}

	// lvisit prepares the solution
	lvisit := func() []int {
		solution := make([]int, n)
		for i := 1; i < n+1; i++ {
			solution[i-1] = (x[i] % 2) ^ 1
		}
		if debug {
			log.Printf("visit solution=%v", solution)
		}

		return solution
	}

	// isUnit determines if literal l is being watched in some clause whose other
	// literals are entirely false; return 0 if false, 1 if true
	isUnit := func(l int) int {
		j := watch[l]

		for j != 0 {
			// step (i)
			p := start[j] + 1

			// step (ii)
		D3ii:
			if p == start[j-1] {
				return 1
			}

			// step (iii) - Check if L(p) is false
			if x[state[p].L>>1] == state[p].L&1 {
				p += 1
				goto D3ii
			}

			// step (iv)
			j = link[j]
		}

		return 0
	}

	//
	// D1 [Initialize.]
	//

	initialize()

	if debug {
		log.Printf("D1. Initialize")
	}

	moves[0] = 0
	d = 0
	head = 0
	tail = 0

	for k = n; k >= 1; k-- {
		x[k] = -1 // an unset value
		if watch[2*k] != 0 || watch[2*k+1] != 0 {
			next[k] = head
			head = k
			if tail == 0 {
				tail = k
			}
		}
	}

	if tail != 0 {
		next[tail] = head
	}

	if debug {
		dump()
	}

	if progress {
		showProgress()
	}

	//
	// D2. [Success?]
	//
D2:
	if tail == 0 {
		// visit the solution
		if debug {
			log.Println("D2. [Success!]")
		}
		if stats != nil {
			stats.Solutions++
		}

		return true, lvisit()
	}

	k = tail

	if stats != nil {
		stats.Levels[d]++
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
	// D3. [Look for unit clauses.]
	//
D3:
	head = next[k]

	if debug {
		log.Printf("D3. [Look for unit clauses.] active=%s, k=%d, head=%d", activeRing(), k, head)
	}

	// Compute f = [2h is a unit] + 2[2h + 1 is a unit]
	f = isUnit(2*head) + 2*isUnit(2*head+1)
	if debug {
		log.Printf("D3. f=%d", f)
	}

	if f == 3 {
		goto D7 // [Backtrack.]
	}

	if f == 1 || f == 2 {
		moves[d+1] = f + 3
		tail = k
		goto D5 // [Move on.]
	}

	if head != tail {
		k = head
		goto D3 // [Look for unit clauses.]
	}

	//
	// D4. [Two-way branch.]
	//

	head = next[tail]
	if watch[2*head] == 0 || watch[2*head+1] != 0 {
		moves[d+1] = 1
	} else {
		moves[d+1] = 0
	}

	if debug {
		log.Printf("D4. [Two-way branch.] d=%d, x=%v, moves=%v", d, x[1:], moves[1:d+1])
	}

D5:
	//
	// D5. [Move on.]
	//
	if debug {
		log.Printf("D5. [Move on.]")
	}

	d += 1
	h[d] = head
	k = head

	if tail == k {
		tail = 0
	} else {
		// delete variable k from the ring
		next[tail] = next[k]
		head = next[k]
	}

	//
	// D6. [Update watches.]
	//
D6:
	if debug {
		log.Printf("D6. [Update watches.]")
	}

	b = (moves[d] + 1) % 2
	x[k] = b

	// Clear the watch list for ^(x_k)
	l = 2*k + b
	j = watch[l]
	watch[l] = 0

	for j != 0 {

		// step (i)
		jp = link[j]
		i = start[j]
		p = i + 1

		// step (ii) - loop while L(p) is false
		// will end before p == start[j-1]
		for x[state[p].L>>1] == state[p].L&1 {
			p += 1
		}

		// step (iii)
		lp = state[p].L
		state[p].L = l
		state[i].L = lp

		// step (iv)
		p = watch[lp]
		q = watch[lp^1]
		if p != 0 || q != 0 || x[lp>>1] >= 0 {
			goto D6vi
		}

		// step (v) - Insert |l'| into the ring as its new head
		if tail == 0 {
			tail = lp >> 1
			head = lp >> 1
			next[tail] = head
		} else {
			next[lp>>1] = head
			head = lp >> 1
			next[tail] = head
		}

		// step (vi)
	D6vi:
		// Insert j into the watch list of l'
		link[j] = p
		watch[lp] = j

		// step (vii)
		j = jp
	}

	if debug {
		log.Printf("D6. d=%d, l=%d, active=%s, x=%v, moves=%v", d, l, activeRing(), x[1:], moves[1:d+1])
	}

	goto D2

	//
	// D7. [Backtrack.]
	//
D7:
	if debug {
		log.Printf("D7. [Backtrack.]")
	}

	tail = k

	for moves[d] >= 2 {
		k = h[d]
		x[k] = -1
		if watch[2*k] != 0 || watch[2*k+1] != 0 {
			next[k] = head
			head = k
			next[tail] = head
		}
		d -= 1
	}

	//
	// D8. [Failure?]
	//
	if debug {
		log.Printf("D8. [Failure?]")
	}

	if d > 0 {
		moves[d] = 3 - moves[d]
		k = h[d]
		goto D6
	}

	// Terminate, the clauses aren't satisfiable
	return false, nil
}
