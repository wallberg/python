package taocp

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

// XCC implements Algorithm C (7.2.2.1), exact covering with colors via
// dancing links.  The task is to find all subsets of options such
// that:
//
// 1) each primary item j occurs exactly once
// 2) every secondary item has been assigned at most one color
//
// Arguments:
// items     -- sorted list of primary items
// options   -- list of list of options; every option must contain at least one
// 			    primary item
// secondary -- sorted list of secondary items; can contain an optional
//              "color" appended after a colon, eg "sitem:color"
// stats     -- structure to capture runtime statistics and provide feedback on
//              progress
// minimax   -- when true, only visit solutions whose maximum option number is
//              <= the maximum option number of any solution already found
// minimaxSingle
//           -- return only one minimax solution for a given maximium option number
// visit     -- function called with each discovered solution, returns true
//              if the search should continue
//
func XCC(items []string, options [][]string, secondary []string,
	stats *ExactCoverStats, minimax bool, minimaxSingle bool,
	visit func(solution [][]string) bool) error {

	var (
		n1       int      // number of primary items
		n2       int      // number of secondary items
		n        int      // total number of items
		size     int      // total size of the options table
		name     []string // name of the item
		llink    []int    // right link of the item
		rlink    []int    // left link of the item
		top      []int    // pointer to the vertical list header (item)
		llen     []int
		ulink    []int
		dlink    []int
		color    []int    // color of a particular item in option
		colors   []string // map of color names, key is the index starting at 1
		level    int
		state    []int // search state
		cutoff   int   // pointer to the spacer at one end of the best minimax solution found so far
		debug    bool  // is debug enabled?
		progress bool  // is progress enabled?
	)

	dump := func() {
		var b strings.Builder
		b.WriteString("\n")

		// Tables
		b.WriteString(fmt.Sprintf("name :  %v\n", name))
		b.WriteString(fmt.Sprintf("llink:  %v\n", llink))
		b.WriteString(fmt.Sprintf("rlink:  %v\n", rlink))
		b.WriteString(fmt.Sprintf("top  :  %v\n", top))
		b.WriteString(fmt.Sprintf("llen :  %v\n", llen))
		b.WriteString(fmt.Sprintf("ulink:  %v\n", ulink))
		b.WriteString(fmt.Sprintf("dlink:  %v\n", dlink))
		b.WriteString(fmt.Sprintf("color:  %v\n", color))
		b.WriteString("colors: ")
		for i, colorName := range colors {
			if i > 0 {
				b.WriteString(fmt.Sprintf(" %d=%s", i, colorName))
			}
		}
		b.WriteString("\n")

		// Remaining items
		b.WriteString("items:  ")
		i := 0
		for rlink[i] != 0 {
			i = rlink[i]
			b.WriteString(" " + name[i])
		}
		b.WriteString("\n")

		// Selected options
		for i, p := range state[0:level] {
			b.WriteString(fmt.Sprintf("  option: i=%d, p=%d (", i, p))
			// Move back to first element in the option
			for top[p-1] > 0 {
				p--
			}
			// Iterate over elements in the option
			q := p
			for top[q] > 0 {
				b.WriteString(fmt.Sprintf(" %v", name[top[q]]))
				q++
			}
			b.WriteString(" )\n")
		}
		// log.Print(b.String())

		//
		// Populate the matrix
		//

		var m [][]string

		// List the items
		m = append(m, make([]string, n+2))
		m[0][0] = "NAME"

		// primary items
		for i := rlink[0]; i != 0; i = rlink[i] {
			m[0][i] = fmt.Sprintf("%d,%s", i, name[i])
		}

		// secondary items
		for i := rlink[n+1]; i != n+1; i = rlink[i] {
			m[0][i] = fmt.Sprintf("%d,%s", i, name[i])
		}

		// m[0][n+1] = fmt.Sprintf("%d -", n+1)

		// spacer row
		m = append(m, make([]string, n+2))

		// llen
		m = append(m, make([]string, n+2))
		m[2][0] = "LLEN"
		for i := 1; i < n+1; i++ {
			m[2][i] = fmt.Sprintf("%d", llen[i])
		}

		// options
		for p := n + 1; p < size-1; {
			var row []string

			// spacer row
			row = make([]string, n+2)
			if p == cutoff {
				row[0] = "CUTOFF"
			}
			m = append(m, row)

			// start a new option
			row = make([]string, n+2)
			row[0] = fmt.Sprintf("%d:%d", p, top[p])

			// add each item in the option
			for p++; top[p] > 0; p++ {
				// only add if the item is still in the list
				i := top[p]
				for q := dlink[i]; q != i; q = dlink[q] {
					if q == p {
						s := name[i]
						if color[p] > 0 {
							s += ":" + colors[color[p]]
						}
						row[top[p]] = fmt.Sprintf("%d,%s", p, s)
						break
					}
				}
			}

			// finish this option
			row[n+1] = fmt.Sprintf("%d:%d", p, top[p])
			m = append(m, row)
		}

		//
		// Get the column widths
		//

		widths := make([]int, n+2)
		for _, row := range m {
			for col := 0; col < n+2; col++ {
				if len(row[col]) > widths[col] {
					widths[col] = len(row[col])
				}
			}
		}

		//
		// Prepare the output
		//

		b.Reset()
		b.WriteString("\n")

		// Add the level
		b.WriteString(fmt.Sprintf("level=%d\n", level))

		// Add the state
		b.WriteString("x=")
		for _, p := range state[0:level] {
			b.WriteString(name[top[p]])
			b.WriteString(" ")
		}
		b.WriteString("\n")

		// Add the matrix
		for _, row := range m {
			for col := 0; col < n+2; col++ {
				b.WriteString(row[col])

				// Add the blank padding
				for p := len(row[col]); p < widths[col]; p++ {
					b.WriteString(" ")
				}
				b.WriteString("  ")

				if col == n1 {
					// Add seperator between primary and secondary items
					b.WriteString(" | ")
				}
			}
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("COLOR:  %v\n", color))

		log.Print(b.String())
	}

	validate := func() error {
		// Items
		if len(items) == 0 {
			return fmt.Errorf("items may not be empty")
		}
		mItems := make(map[string]bool)
		for _, item := range items {
			if mItems[item] {
				return fmt.Errorf("Item '%s' is not unique", item)
			}
			mItems[item] = true
		}

		// Secondary Items
		mSItems := make(map[string]bool)
		for _, sitem := range secondary {
			if mItems[sitem] || mSItems[sitem] {
				return fmt.Errorf("Secondary item '%s' is not unique", sitem)
			}
			mSItems[sitem] = true
		}

		// Options
		for _, option := range options {
			for _, item := range option {
				i := strings.Index(item, ":")
				if i > -1 {
					item = item[:i]
				}
				if !mItems[item] && !mSItems[item] {
					return fmt.Errorf("Option '%v' contains '%s' which is not an item or secondary item", option, item)
				}
			}
		}

		return nil
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
			progress = stats.Progress
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

		// Fill out the option tables
		nOptions := len(options)
		nOptionItems := 0
		for _, option := range options {
			nOptionItems += len(option)
		}
		size = n + 1 + nOptions + 1 + nOptionItems

		top = make([]int, size)
		llen = top[0 : n+1] // first n+1 elements of top
		ulink = make([]int, size)
		dlink = make([]int, size)
		color = make([]int, size)
		colors = make([]string, 1)

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

				// Extract the color
				itemColor := 0 // 0 if there is no color
				iColon := strings.Index(item, ":")
				if iColon > -1 {
					itemColorName := item[iColon+1:]
					item = item[:iColon]

					// Insert the color name into color[]
					for itemColor = 1; itemColor < len(colors); itemColor++ {
						if itemColorName == colors[itemColor] {
							break
						}
					}
					if itemColor == len(colors) {
						// Not found, add new color name entry
						colors = append(colors, itemColorName)
					}
				}

				// Insert the item into name[]
				i := 0
				for _, value := range name {
					if value == item {
						break
					}
					i++
				}
				top[x] = i
				color[x] = itemColor

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

		level = 0
		state = make([]int, nOptions)
		cutoff = math.MaxInt16

		if debug {
			dump()
		}
	}

	showProgress := func() {

		if debug && stats.Verbosity > 0 {
			dump()
		}

		est := 0.0 // estimate of percentage done
		tcum := 1

		var b strings.Builder
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Current level %d of max %d\n", level, stats.MaxLevel))

		// Iterate over the options
		for _, p := range state[0:level] {
			// Cyclically gather the items in the option, beginning at p
			q := p
			b.WriteString(" ")
			for {
				b.WriteString(fmt.Sprintf(" %v", name[top[q]]))
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
				b.WriteString(" not in this list\n")
			}
		}

		est += 1.0 / float64(2*tcum)

		b.WriteString(fmt.Sprintf("solutions=%d, nodes=%d, est=%4.4f\n",
			stats.Solutions, stats.Nodes, est))
		log.Print(b.String())
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

	// hide removes an option from further consideration
	hide := func(p int) {
		if debug && stats.Verbosity > 1 {
			log.Printf("hide(p=%d)", p)
		}

		// iterate over the items in this option
		q := p + 1
		for q != p {
			x := top[q]
			u, d := ulink[q], dlink[q]
			if x <= 0 {
				q = u // q was a spacer
			} else {
				if color[q] >= 0 {
					dlink[u], ulink[d] = d, u
					llen[x]--
				}
				q++
			}
		}
	}

	unhide := func(p int) {
		if debug && stats.Verbosity > 1 {
			log.Printf("unhide(p=%d)", p)
		}

		q := p - 1
		// log.Printf("q=%d", q)
		for q != p {
			x := top[q]
			// log.Printf("x=%d", x)
			u, d := ulink[q], dlink[q]
			// log.Printf("u=%d, d=%d", u, d)
			if x <= 0 {
				q = d // q was a spacer
			} else {
				if minimax && d > cutoff {
					// d = x
					// dlink[q] = x
					dlink[q], d = x, x
				}
				if color[q] >= 0 {
					dlink[u], ulink[d] = q, q
					llen[x]++
				}
				q--
			}
			// log.Printf("q=%d", q)
		}
	}

	// cover removes i from the list of items needing to be covered and
	// hides all of the item's options
	cover := func(i int) {
		if debug && stats.Verbosity > 1 {
			log.Printf("cover(i=%d)", i)
		}

		// hide all of the item's options
		p := dlink[i]
		for p != i {
			hide(p)
			p = dlink[p]
		}
		// remove the item from the list
		l, r := llink[i], rlink[i]
		rlink[l], llink[r] = r, l
	}

	uncover := func(i int) {
		if debug && stats.Verbosity > 1 {
			log.Printf("uncover(i=%d)", i)
		}

		if minimax {
			// Remove all nodes > cutoff from item i's list
			q := dlink[i]
			for q != i {
				if q > cutoff {
					// Remove q from i's list
					u, d := ulink[q], dlink[q]
					dlink[u], ulink[d] = d, u
					llen[i]--
				}
				q = dlink[q]
			}
		}

		l, r := llink[i], rlink[i]
		rlink[l], llink[r] = i, i
		p := ulink[i]
		for p != i {
			unhide(p)
			p = ulink[p]
		}
	}

	purify := func(p int) {
		if debug && stats.Verbosity > 1 {
			log.Printf("purify(p=%d)", p)
		}

		c := color[p]
		i := top[p]
		color[i] = c
		q := dlink[i]
		for q != i {
			if color[q] == c {
				// this secondary item has the same color, the -1 value is a
				// flag that indicates this item is already a match, no need to
				// check the color again.
				color[q] = -1
			} else {
				// this secondary item does not have the same color, so hide it
				hide(q)
			}
			q = dlink[q]
		}
	}

	unpurify := func(p int) {
		if debug && stats.Verbosity > 1 {
			log.Printf("unpurify(p=%d)", p)
		}

		// TODO: determine what we need to do here for minimax
		c := color[p]
		i := top[p]
		q := ulink[i]
		// log.Printf("c=%d, i=%d, q=%d", c, i, q)
		for q != i {
			if color[q] < 0 {
				// Restore the original color, before purification
				color[q] = c
			} else {
				// dump()
				unhide(q)
			}
			q = ulink[q]
			if minimax && q > cutoff {
				dlink[q] = i
				ulink[i] = q
			}
			// log.Printf("q=%d", q)
		}
	}

	commit := func(p int, j int) {
		if debug && stats.Verbosity > 1 {
			log.Printf("commit(p=%d, j=%d)", p, j)
		}

		if color[p] == 0 {
			cover(j)
		}
		if color[p] > 0 {
			purify(p)
		}
	}

	uncommit := func(p int, j int) {
		if debug && stats.Verbosity > 1 {
			log.Printf("uncommit(p=%d, j=%d)", p, j)
		}

		if color[p] == 0 {
			uncover(j)
		}
		if color[p] > 0 {
			unpurify(p)
		}
	}

	lvisit := func() bool {

		pMax := 0 // Track max p for minimax
		lMax := 0 // level for max p

		// Only one of the secondary items will have it's color value, the
		// others will have -1. Save the color and add it to all the matching
		// secondary items at the end.
		sitemColor := make(map[string]string)

		// Iterate over the options
		options := make([][]string, 0)
		for i, p := range state[0:level] {
			if p > pMax {
				pMax = p
				lMax = i
			}
			options = append(options, make([]string, 0))
			// Move back to first element in the option
			for top[p-1] > 0 {
				p--
			}
			// Iterate over elements in the option
			q := p
			for top[q] > 0 {
				name := name[top[q]]
				if color[q] > 0 {
					sitemColor[name] = colors[color[q]]
				}
				options[i] = append(options[i], name)
				q++
			}
		}

		// Add the secondary item colors
		for i, option := range options {
			for j, item := range option {
				if color, ok := sitemColor[item]; ok {
					options[i][j] += ":" + color
				}
			}
		}

		// For minimax, remove all nodes > cutoff (new value)
		if minimax {
			// Find spacer at the end of the option for max x_k
			// For minimaxSingle=true find the spacer before the
			// solution, otherwise the spacer after the solutioin
			pp := pMax
			for top[pp] > 0 {
				if minimaxSingle {
					pp--
				} else {
					pp++
				}
			}

			// If we have new cutoff value, remove all nodes > cutoff from
			// further consideration
			if pp != cutoff {
				cutoff = pp
				for _, p := range state[0:level] {
					x := top[p]
					q := p
					for q != x {
						if q > cutoff {
							// Remove q from the x's list
							u, d := ulink[q], dlink[q]
							dlink[u], ulink[d] = d, u
							// TODO: determine if we really need to process
							// each removed option in order to decrement
							// llen[x]
							llen[x]--
						}
						q = dlink[q]
					}
				}

				// Backtrack for each item in state >= lMax
				if minimaxSingle {
					for k := level - 1; k >= lMax; k-- {
						i := top[state[k]]
						uncover(i)
						level--
					}
				}
			}
		}

		if debug {
			log.Printf("visit(%v)", options)
		}

		return visit(options)
	}

	// C1 [Initialize.]
	if stats != nil && stats.Debug {
		log.Printf("C1. Initialize")
	}

	if err := validate(); err != nil {
		return err
	}
	initialize()

	var (
		i int
		j int
		p int
	)

	if progress {
		showProgress()
	}

C2:
	// C2. [Enter level l.]
	if debug {
		log.Printf("C2. l=%d, x[0:l]=%v\n", level, state[0:level])
	}

	if stats != nil {
		stats.Levels[level]++
		stats.Nodes++

		if progress {
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
			log.Println("C2. Visit the solution")
		}
		if stats != nil {
			stats.Solutions++
		}
		resume := lvisit()
		if !resume {
			if debug {
				log.Println("C2. Halting the search")
			}
			if progress {
				showProgress()
			}
			return nil
		}
		goto C8
	}

	// C3. [Choose i.]
	i = mrv()

	if debug {
		log.Printf("C3. Choose i=%d (%s)\n", i, name[i])
	}

	// C4. [Cover i.]
	if debug {
		log.Printf("C4. Cover i=%d (%s)\n", i, name[i])
	}
	cover(i)
	state[level] = dlink[i]

C5:
	// C5. [Try x_l.]
	if debug {
		log.Printf("C5. Try l=%d, x[0:l+1]=%v\n", level, state[0:level+1])
	}
	if state[level] == i {
		goto C7
	}
	// Commit each of the items in this option
	p = state[level] + 1
	for p != state[level] {
		j := top[p]
		if j <= 0 {
			// spacer, go back to the first option
			p = ulink[p]
		} else {
			commit(p, j)
			p++
		}
	}
	level++
	goto C2

C6:
	// C6. [Try again.]
	if debug {
		log.Println("C6. Try again")
	}

	if stats != nil {
		stats.Nodes++
	}

	// Uncommit each of the items in this option
	p = state[level] - 1
	for p != state[level] {
		j = top[p]
		if j <= 0 {
			p = dlink[p]
		} else {
			uncommit(p, j)
			p--
		}
	}

	// Exercise 7.2.2.1-83 performance improvement
	if level == 0 {
		x := state[0]
		for ; top[x] > 0; x++ {
		}
		j = top[x-1]
		if j > n && color[x-1] == 0 {
			cover(j)
		}
	}

	i = top[state[level]]
	state[level] = dlink[state[level]]
	goto C5

C7:
	// C7. [Backtrack.]
	if debug {
		log.Println("C7. Backtrack")
	}
	uncover(i)

C8:
	// C8. [Leave level l.]
	if debug {
		log.Printf("C8. Leaving level %d\n", level)
	}
	if level == 0 {
		if progress {
			showProgress()
		}
		return nil
	}
	level--
	goto C6
}

// Exercise 7.2.2.1-66 Construct sudoku puzzles by placing nine given cars in a
// 3x3 array

// SudokuCards constructs sudoku puzzles with one solution, given nine 3x3
// cards to order. Returns both the card ordering and the matching SuDoku grid.
func SudokuCards(cards [9][3][3]int, stats *ExactCoverStats,
	visit func(cards [9]int, grid [9][9]int) bool) {

	var (
		i int // row number (0-8)
		j int // column number (0-8)
		k int // cell value in (row,column)
		x int // 3x3 box (0-8), aka card slot
		c int // card number (1-9)
	)

	// Build one [p, r, c, b] option
	buildOption := func() []string {
		return []string{
			fmt.Sprintf("p%d%d", i, j),      // piece
			fmt.Sprintf("r%d%d", i, k),      // piece in row
			fmt.Sprintf("c%d%d", j, k),      // piece in column
			fmt.Sprintf("x%d%d", x, k),      // piece in 3x3 box
			fmt.Sprintf("%d%d:%d", i, j, k), // (row, column) with color k
		}
	}

	// Build the items, secondary items, and options
	itemSet := make(map[string]bool)  // set of primary items
	sitemSet := make(map[string]bool) // secondary items (row, column)
	options := make([][]string, 0)

	// Placements within the grid
	for i = 0; i < 9; i++ {
		for j = 0; j < 9; j++ {
			sitemSet[fmt.Sprintf("%d%d", i, j)] = true
			x = 3*(i/3) + (j / 3)
			for k = 1; k < 10; k++ {
				option := buildOption()
				for _, item := range option[:4] {
					itemSet[item] = true
				}
				options = append(options, option)
			}
		}
	}

	// each card is an item
	for c = 1; c <= 9; c++ {
		itemSet[strconv.Itoa(c)] = true
	}

	// each slot is an item
	for x = 0; x < 9; x++ {
		itemSet[fmt.Sprintf("s%d", x)] = true
	}

	// Create one option for of 9 cards in each of 9 slots. Each card ordering
	// has 3!3! symmetric orderings which produce identical results, so use
	// ordering constraints to produce only the first ordering:
	// - Put card 1 in slot 0
	// - Ensure the card in slot 4 is less than the cards in slots 5, 7, and 8

	for c = 1; c <= 9; c++ {
		for x = 0; x < 9 && !(c == 1 && x > 0); x++ {
			option := []string{strconv.Itoa(c), fmt.Sprintf("s%d", x)}
			// Iterate over values in this card
			for iCard := 0; iCard < 3; iCard++ {
				for jCard := 0; jCard < 3; jCard++ {
					k = cards[c-1][iCard][jCard]
					if k > 0 {
						i, j := (x/3)*3, (x%3)*3
						option = append(option, fmt.Sprintf("%d%d:%d",
							i+iCard, j+jCard, k))
					}
				}
			}

			// secondary items which control ordering for 4 < 5,7,8
			if x == 4 {
				t := (c - 1)
				for t > 0 {
					for _, orderX := range []int{5, 7, 8} {
						ord := fmt.Sprintf("o%d%d", orderX, t)
						sitemSet[ord] = true
						option = append(option, ord)
					}
					t = t & (t - 1)
				}

			} else if x == 5 || x == 7 || x == 8 {
				t := -1 - (c - 1)
				for t > -9 {
					ord := fmt.Sprintf("o%d%d", x, -t)
					sitemSet[ord] = true
					option = append(option, ord)
					t = t & (t - 1)
				}
			}

			options = append(options, option)
		}
	}

	// Convert itemSet to a items list
	items := make([]string, len(itemSet))
	i = 0
	for item := range itemSet {
		items[i] = item
		i++
	}
	sort.Strings(items)

	// Convert sitemSet to a sitems list
	sitems := make([]string, len(sitemSet))
	i = 0
	for sitem := range sitemSet {
		sitems[i] = sitem
		i++
	}
	sort.Strings(sitems)

	// Save the accumulated solutions, key=card, value = slice of SuDoku grids
	solutions := make(map[[9]int][][9][9]int)

	// Solve using XCC
	XCC(items, options, sitems, stats, false, false,
		func(solution [][]string) bool {
			var (
				cards [9]int
				grid  [9][9]int
			)

			for _, option := range solution {
				if option[0][0:1] == "p" {
					// SuDoku square
					i, _ := strconv.Atoi(option[4][0:1])
					j, _ := strconv.Atoi(option[4][1:2])
					k, _ := strconv.Atoi(option[4][3:4])
					grid[i][j] = k
				} else {
					// Card
					c, _ := strconv.Atoi(option[0])
					s, _ := strconv.Atoi(option[1][1:2])
					cards[s] = c
				}
			}

			// Add grid to the list for this card ordering
			if grids, ok := solutions[cards]; ok {
				solutions[cards] = append(grids, grid)
			} else {
				solutions[cards] = [][9][9]int{grid}
			}

			return true
		})

	// Return all the card orderings which have one SuDoku grid
	for cards, grids := range solutions {
		if len(grids) == 1 {
			if !visit(cards, grids[0]) {
				break
			}
		}
	}
}

// Mathematicians lists 27 people (without special characters) who were authors
// of early papers in Acta Mathematica and subsequently cited in TAOCP
var Mathematicians = []string{
	"ABEL",
	"BERTRAND",
	"BOREL",
	"CANTOR",
	"CATALAN",
	"FROBENIUS",
	"GLAISHER",
	"GRAM",
	"HADAMARD",
	"HENSEL",
	"HERMITE",
	"HILBERT",
	"HURWITZ",
	"JENSEN",
	"KIRCHHOFF",
	"KNOPP",
	"LANDAU",
	"MARKOFF",
	"MELLIN",
	"MINKOWSKI",
	"NETTO",
	"PERRON",
	"RUNGE",
	"STERN",
	"STIELTJES",
	"SYLVESTER",
	"WEIERSTRASS",
}

// WordSearch uses ExactCoverColoring to build a m x n word search, given the
// provided words
func WordSearch(m int, n int, words []string, stats *ExactCoverStats,
	visit func([][]string) bool) {

	coord := func(i int, j int) string {
		return fmt.Sprintf("%02d%02d", i, j)
	}

	// secondary items
	secondary := make([]string, m*n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			secondary[i*n+j] = coord(i, j)
		}
	}

	// options
	options := make([][]string, 0)
	for x, word := range words {

		for i := 0; i < m; i++ {
			for j := 0; j < n; j++ {
				// Eight directions for each starting position (i,j)
				var wordDs [8][]string // word directions
				for d := 0; d < 8; d++ {
					wordDs[d] = []string{word}
				}

				for k := 0; k < len(word); k++ {
					// right
					if j+k < n {
						wordDs[0] = append(wordDs[0], coord(i, j+k)+":"+word[k:k+1])
					}
					// right-down
					if i+k < m && j+k < n {
						wordDs[1] = append(wordDs[1], coord(i+k, j+k)+":"+word[k:k+1])
					}

					// To avoid symmetric positions, only allow 2 (of 8) directions
					// for the first word
					if x == 0 {
						continue
					}

					// down
					if i+k < m {
						wordDs[2] = append(wordDs[2], coord(i+k, j)+":"+word[k:k+1])
					}
					// left-down
					if i+k < m && j-k >= 0 {
						wordDs[3] = append(wordDs[3], coord(i+k, j-k)+":"+word[k:k+1])
					}
					// left
					if j-k >= 0 {
						wordDs[4] = append(wordDs[4], coord(i, j-k)+":"+word[k:k+1])
					}
					// left-up
					if i-k >= 0 && j-k >= 0 {
						wordDs[5] = append(wordDs[5], coord(i-k, j-k)+":"+word[k:k+1])
					}
					// up
					if i-k >= 0 {
						wordDs[6] = append(wordDs[6], coord(i-k, j)+":"+word[k:k+1])
					}
					// right-up
					if i-k >= 0 && j+k < n {
						wordDs[7] = append(wordDs[7], coord(i-k, j+k)+":"+word[k:k+1])
					}

				}

				for _, wordD := range wordDs {
					if len(wordD) == len(word)+1 {
						options = append(options, wordD)
						// log.Println(wordD)
					}
				}
			}
		}
	}

	XCC(words, options, secondary, stats, false, false,
		func(solution [][]string) bool {
			return visit(solution)
		})
}
