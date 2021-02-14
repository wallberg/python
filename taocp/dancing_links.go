package taocp

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Explore Dancing Links from The Art of Computer Programming, Volume 4,
// Fascicle 5, Mathematical Preliminaries Redux; Introduction to Backtracking;
// Dancing Links, 2020
//
// §7.2.2.1 Dancing Links

// ExactCoverStats is a struct for tracking ExactCover statistics and reporting
// runtime progress
type ExactCoverStats struct {
	// Input parameters
	Progress  bool // Display runtime progress
	Debug     bool // Enable debug logging
	Verbosity int  // Debug verbosity level (0 or 1)
	Delta     int  // Display progress every Delta number of Nodes

	// Statistics collectors
	MaxLevel  int   // Maximum level reached
	Theta     int   // Display progress at next Theta number of Nodes
	Levels    []int // Count of times each level is entered
	Nodes     int   // Count of nodes processed
	Solutions int   // Count of solutions returned
}

func (s ExactCoverStats) String() string {
	// Find first non-zero level count
	i := len(s.Levels)
	for s.Levels[i-1] == 0 && i > 0 {
		i--
	}

	return fmt.Sprintf("nodes=%d, solutions=%d, levels=%v", s.Nodes,
		s.Solutions, s.Levels[:i])
}

// ExactCoverYaml provides YAML (de-)serialization for Exact Cover input
type ExactCoverYaml struct {
	Items   []string `yaml:""` // Primary Items
	SItems  []string `yaml:""` // Secondary Items
	Options []string `yaml:""` // Options
}

// NewExactCoverYaml creates a new instance of ExactCoverYaml
func NewExactCoverYaml(items []string, sitems []string, options [][]string) *ExactCoverYaml {
	xcYaml := ExactCoverYaml{Items: items, SItems: sitems}
	xcYaml.Options = make([]string, 0)
	for _, option := range options {
		xcYaml.Options = append(xcYaml.Options,
			strings.Join(option, " "))
	}

	return &xcYaml
}

// ExactCover implements Algorithm X, exact cover via dancing links.
// The task is to find all subsets of options such that each primary item j
// occurs exactly once.
//
// Arguments:
// items     -- sorted list of primary items
// options   -- list of list of options; every option must contain at least one
// 			    primary item
// secondary -- sorted list of secondary items
// stats     -- structure to capture runtime statistics and provide feedback on
//              progress
// visit     -- function called with each discovered solution, returns true
//              if the search should resume
//
func ExactCover(items []string, options [][]string, secondary []string,
	stats *ExactCoverStats, visit func(solution [][]string) bool) {

	var (
		n1    int      // number of primary items
		n2    int      // number of secondary items
		n     int      // total number of items
		name  []string // name of the item
		llink []int    // right link of the item
		rlink []int    // left link of the item
		top   []int
		llen  []int
		ulink []int
		dlink []int
		level int
		state []int // search state
		debug bool  // is debug enabled?
	)

	dump := func() {
		i := 0
		for rlink[i] != 0 {
			i = rlink[i]
			log.Print("  ", name[i])
			x := i
			for dlink[x] != i {
				x = dlink[x]
				log.Print(" ", name[top[x]])
			}
			log.Println()
		}
		log.Println("---")
	}

	initialize := func() {
		n1 = len(items)
		n2 = len(secondary)
		n = n1 + n2

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
		}

		// Fill out the item tables
		name = make([]string, n+2)
		llink = make([]int, n+2)
		rlink = make([]int, n+2)

		for j, item := range append(items, secondary...) {
			i := j + 1
			name[i] = item
			llink[i] = i - 1
			rlink[i-1] = i
		}

		// two doubly linked lists, primary and secondary
		// head of the primary list is at i=0
		// head of the secondary list is at i=n+1
		llink[n+1] = n
		rlink[n] = n + 1
		llink[n1+1] = n + 1
		rlink[n+1] = n1 + 1
		llink[0] = n1
		rlink[n1] = 0

		if debug {
			log.Println("name", name)
			log.Println("llink", llink)
			log.Println("rlink", rlink)
		}

		// Fill out the option tables
		nOptions := len(options)
		nOptionItems := 0
		for _, option := range options {
			nOptionItems += len(option)
		}
		size := n + 1 + nOptions + 1 + nOptionItems

		top = make([]int, size)
		llen = top[0 : n+1] // first n+1 elements of top
		ulink = make([]int, size)
		dlink = make([]int, size)

		// Set empty list for each item
		for i := 1; i <= n; i++ {
			llen[i] = 0
			ulink[i] = i
			dlink[i] = i
		}

		// Insert each of the options and their items
		x := n + 1
		spacer := 0
		top[x] = spacer
		spacerX := x

		// Iterate over each option
		for _, option := range options {
			// Iterate over each item in this option
			for _, item := range option {
				x++
				i := 0
				for _, value := range name {
					if value == item {
						break
					}
					i++
				}
				top[x] = i

				// Insert into the option list for this item
				llen[i]++ // increase the size by one
				head := i
				tail := i
				for dlink[tail] != head {
					tail = dlink[tail]
				}

				dlink[tail] = x
				ulink[x] = tail

				ulink[head] = x
				dlink[x] = head
			}

			// Insert spacer at end of each option
			dlink[spacerX] = x
			x++
			ulink[x] = spacerX + 1

			spacer--
			top[x] = spacer
			spacerX = x
		}

		if debug {
			log.Println("top", top)
			log.Println("llen", llen)
			log.Println("ulink", ulink)
			log.Println("dlink", dlink)
		}

		level = 0
		state = make([]int, nOptions)

		if debug {
			dump()
		}
	}

	showProgress := func() {

		est := 0.0 // estimate of percentage done
		tcum := 1

		log.Printf("Current level %d of max %d\n", level, stats.MaxLevel)

		// Iterate over the options
		for _, p := range state[0:level] {
			// Cyclically gather the items in the option, beginning at p
			var b strings.Builder
			b.WriteString("  ")
			q := p
			for {
				b.WriteString(name[top[q]] + " ")
				q++
				if top[q] <= 0 {
					q = ulink[q]
				}
				if q == p {
					break
				}
			}

			// Get position stats for this option
			i := top[p]
			q = dlink[i]
			k := 1
			for q != p && q != i {
				q = dlink[q]
				k++
			}

			if q != i {
				b.WriteString(fmt.Sprintf(" %d of %d\n", k, llen[i]))
				tcum *= llen[i]
				est += float64(k-1) / float64(tcum)
			} else {
				b.WriteString(" not in this list")
			}
			log.Print(b.String())
		}

		est += 1.0 / float64(2*tcum)

		log.Printf("  solutions=%d, nodes=%d, est=%4.4f\n",
			stats.Solutions, stats.Nodes, est)
		log.Println("---")
	}

	lvisit := func() bool {
		// Iterate over the options
		options := make([][]string, 0)
		for i, p := range state[0:level] {
			options = append(options, make([]string, 0))
			// Move back to first element in the option
			for top[p-1] > 0 {
				p--
			}
			// Iterate over elements in the option
			q := p
			for top[q] > 0 {
				options[i] = append(options[i], name[top[q]])
				q++
			}
		}

		return visit(options)
	}

	// mrv selects the next item to try using the Minimum Remaining
	// Values heuristic.
	mrv := func() int {

		i := 0
		theta := -1
		p := rlink[0]
		for p != 0 {
			lambda := llen[p]
			if lambda < theta || theta == -1 {
				theta = lambda
				i = p
				if theta == 0 {
					return i
				}
			}
			p = rlink[p]
		}

		return i
	}

	hide := func(p int) {
		q := p + 1
		for q != p {
			x := top[q]
			u, d := ulink[q], dlink[q]
			if x <= 0 {
				q = u // q was a spacer
			} else {
				dlink[u], ulink[d] = d, u
				llen[x]--
				q++
			}
		}
	}

	cover := func(i int) {
		p := dlink[i]
		for p != i {
			hide(p)
			p = dlink[p]
		}
		l, r := llink[i], rlink[i]
		rlink[l], llink[r] = r, l
	}

	unhide := func(p int) {
		q := p - 1
		for q != p {
			x := top[q]
			u, d := ulink[q], dlink[q]
			if x <= 0 {
				q = d // q was a spacer
			} else {
				dlink[u], ulink[d] = q, q
				llen[x]++
				q--
			}
		}
	}

	uncover := func(i int) {
		l, r := llink[i], rlink[i]
		rlink[l], llink[r] = i, i
		p := ulink[i]
		for p != i {
			unhide(p)
			p = ulink[p]
		}
	}

	// X1 [Initialize.]
	initialize()

	var (
		i int
		j int
		p int
	)

X2:
	// X2. [Enter level l.]
	if debug {
		log.Printf("X2. level=%d, x=%v\n", level, state[0:level])
	}

	if stats != nil {
		stats.Levels[level]++
		stats.Nodes++

		if stats.Progress {
			if level > stats.MaxLevel {
				stats.MaxLevel = level
			}
			if stats.Nodes >= stats.Theta {
				showProgress()
				stats.Theta += stats.Delta
			}
		}
	}

	if rlink[0] == 0 {
		// visit the solution
		if debug {
			log.Println("X2. Visit the solution")
		}
		if stats != nil {
			stats.Solutions++
		}
		resume := lvisit()
		if !resume {
			if debug {
				log.Println("X2. Halting the search")
			}
			return
		}
		goto X8
	}

	// X3. [Choose i.]
	i = mrv()

	if debug {
		log.Printf("X3. Choose i=%d (%s)\n", i, name[i])
	}

	// X4. [Cover i.]
	if debug {
		log.Printf("X4. Cover i=%d (%s)\n", i, name[i])
	}
	cover(i)
	state[level] = dlink[i]

X5:
	// X5. [Try x_l.]
	if debug {
		log.Printf("X5. Try l=%d, x[l]=%d\n", level, state[level])
	}
	if state[level] == i {
		goto X7
	}
	p = state[level] + 1
	for p != state[level] {
		j := top[p]
		if j <= 0 {
			p = ulink[p]
		} else {
			cover(j)
			p++
		}
	}
	level++
	goto X2

X6:
	// X6. [Try again.]
	if debug {
		log.Println("X6. Try again")
	}

	if stats != nil {
		stats.Nodes++
	}

	p = state[level] - 1
	for p != state[level] {
		j = top[p]
		if j <= 0 {
			p = dlink[p]
		} else {
			uncover(j)
			p--
		}
	}
	i = top[state[level]]
	state[level] = dlink[state[level]]
	goto X5

X7:
	// X7. [Backtrack.]
	if debug {
		log.Println("X7. Backtrack")
	}
	uncover(i)

X8:
	// X8. [Leave level l.]
	if debug {
		log.Printf("X8. Leaving level %d\n", level)
	}
	if level == 0 {
		return
	}
	level--
	goto X6
}

// LangfordPairs uses ExactCover to return solutions for Langford pairs
// of n values
func LangfordPairs(n int, stats *ExactCoverStats, visit func(solution []int) bool) {

	// Build the list of items
	items := make([]string, 3*n)
	for i := 0; i < n; i++ {
		items[i] = strconv.Itoa(i + 1)
	}
	for i := 0; i < 2*n; i++ {
		items[n+i] = "s" + strconv.Itoa(i)
	}

	// Build the list of options
	options := make([][]string, 0)
	for i := 1; i <= n; i++ {
		j := 1
		k := j + i + 1
		for k <= 2*n {
			// Exercise 15: Omit the reversals
			x := 0
			if n%2 == 0 {
				x = 1
			}
			if i != n-x || j <= n/2 {
				options = append(options,
					[]string{
						strconv.Itoa(i),
						"s" + strconv.Itoa(j-1),
						"s" + strconv.Itoa(k-1),
					})
			}
			j++
			k++
		}
	}

	// Generate solutions
	ExactCover(items, options, []string{}, stats,
		func(solution [][]string) bool {
			x := make([]int, 2*n)
			for _, option := range solution {
				value, _ := strconv.Atoi(option[0])
				i, _ := strconv.Atoi(string(option[1][1]))
				j, _ := strconv.Atoi(string(option[2][1]))

				x[i], x[j] = value, value
			}
			return visit(x)
		})
}

// NQueens uses ExactCover to return solutions for the n-queens problem
func NQueens(n int, stats *ExactCoverStats, visit func(solution []string) bool) {

	items := make([]string, 2*n)
	sitems := make([]string, 4*n-2)
	options := make([][]string, n*n)

	k := 0
	for i := 0; i < n; i++ {
		row := "r" + strconv.Itoa(i+1)
		items[i] = row
		for j := 0; j < n; j++ {
			col := "c" + strconv.Itoa(j+1)
			if i == n-1 {
				items[j+n] = col
			}
			upDiag := "a" + strconv.Itoa(i+j+2)
			downDiag := "b" + strconv.Itoa(i-j)

			var x int

			for x = 0; sitems[x] != upDiag && sitems[x] != ""; x++ {
			}
			if sitems[x] == "" {
				sitems[x] = upDiag
			}

			for x = 0; sitems[x] != downDiag && sitems[x] != ""; x++ {
			}
			if sitems[x] == "" {
				sitems[x] = downDiag
			}

			options[k] = []string{row, col, upDiag, downDiag}
			k++
		}
	}

	// Generate solutions
	ExactCover(items, options, sitems, stats,
		func(solution [][]string) bool {
			x := make([]string, 2*n)
			i := 0
			for _, option := range solution {
				x[i] = option[0]   // row
				x[i+1] = option[1] // col
				i += 2
			}
			return visit(x)
		})
}

// Sudoku uses ExactCover to solve 9x9 sudoku puzzles
func Sudoku(grid [9][9]int, stats *ExactCoverStats,
	visit func(solution [9][9]int) bool) {

	var (
		i int // row number
		j int // column number
		k int // cell value in (row,column)
		x int // 3x3 box
	)

	// Build the [p, r, c, b] option
	buildOption := func() []string {
		return []string{
			"p" + strconv.Itoa(i) + strconv.Itoa(j), // piece
			"r" + strconv.Itoa(i) + strconv.Itoa(k), // piece in row
			"c" + strconv.Itoa(j) + strconv.Itoa(k), // piece in column
			"x" + strconv.Itoa(x) + strconv.Itoa(k), // piece in 3x3 box
		}
	}

	// Get the known items (non zero) provided in the grid
	knownItems := make(map[string]bool)
	for i = 0; i < 9; i++ {
		for j = 0; j < 9; j++ {
			k = grid[i][j]
			if k > 0 {
				x = 3*(i/3) + (j / 3)
				for _, item := range buildOption() {
					knownItems[item] = true
				}
			}
		}
	}

	// Build the items and options from the unknown values
	itemSet := make(map[string]bool)
	options := make([][]string, 0)
	for i = 0; i < 9; i++ {
		for j = 0; j < 9; j++ {
			x = 3*(i/3) + (j / 3)
			for k = 1; k < 10; k++ {
				option := buildOption()
				if !(knownItems[option[0]] || knownItems[option[1]] ||
					knownItems[option[2]] || knownItems[option[3]]) {
					for _, item := range option {
						itemSet[item] = true
					}
					options = append(options, option)
				}
			}
		}
	}

	items := make([]string, len(itemSet))
	i = 0
	for item := range itemSet {
		items[i] = item
		i++
	}

	// Generate solutions
	ExactCover(items, options, []string{}, stats,
		func(solution [][]string) bool {
			// Make a copy of the original grid
			var x [9][9]int
			for i := 0; i < 9; i++ {
				for j := 0; j < 9; j++ {
					x[i][j] = grid[i][j]
				}
			}

			// Fill in the solution values
			for _, option := range solution {
				i, _ := strconv.Atoi(string(option[0][1]))
				j, _ := strconv.Atoi(string(option[0][2]))
				k, _ := strconv.Atoi(string(option[1][2]))

				x[i][j] = k
			}

			return visit(x)
		})
}
