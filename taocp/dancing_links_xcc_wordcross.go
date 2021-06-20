package taocp

import (
	"fmt"
	"log"
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
// xccOptions
//           -- various processing options for XCC; nil value is equivalent to
//              &XCCOptions{} with all default values
//
// visit     -- function called with each discovered solution, returns true
//              if the search should continue
//
func XCCWordCross(items []string, options [][]string, secondary []string,
	stats *ExactCoverStats, xccOptions *XCCOptions,
	visit func(solution [][]string) bool) error {

	if xccOptions == nil {
		// Use all default values
		xccOptions = &XCCOptions{}
	}

	var (
		n1       int      // number of primary items
		n2       int      // number of secondary items
		n        int      // total number of items (N)
		m        int      // total number of options (M)
		size     int      // total size of the options table
		name     []string // name of the item
		rlink    []int    // right link of the item
		llink    []int    // left link of the item
		top      []int    // pointer to the vertical list header (item)
		llen     []int
		ulink    []int
		dlink    []int
		color    []int    // color of a particular item in option
		colors   []string // map of color names, key is the index starting at 1
		level    int
		state    []int  // search state
		xcBranch []bool // At each level, are we branching on connected words instead of next_item?
		cutoff   int    // pointer to the spacer at one end of the best minimax solution found so far
		debug    bool   // is debug enabled?
		progress bool   // is progress enabled?
	)

	dump := func() {

		if stats.SuppressDump {
			return
		}

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
				return fmt.Errorf("item '%s' is not unique", item)
			}
			mItems[item] = true
		}

		// Secondary Items
		mSItems := make(map[string]bool)
		for _, sitem := range secondary {
			if mItems[sitem] || mSItems[sitem] {
				return fmt.Errorf("secondary item '%s' is not unique", sitem)
			}
			mSItems[sitem] = true
		}

		// Options
		if len(options) == 0 {
			return fmt.Errorf("options may not be empty")
		}
		for _, option := range options {
			for _, item := range option {
				i := strings.Index(item, ":")
				if i > -1 {
					item = item[:i]
				}
				if !mItems[item] && !mSItems[item] {
					return fmt.Errorf("option '%v' contains '%s' which is not an item or secondary item", option, item)
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
		m = len(options)
		nOptionItems := 0
		for _, option := range options {
			nOptionItems += len(option)
		}
		size = n + 1 + m + 1 + nOptionItems

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

				// Insert into the option list for this item
				var i int
				var value string
				for i, value = range name {
					if value == item {
						break
					}
				}

				top[x] = i
				color[x] = itemColor

				llen[i]++ // increase the size by one

				head := i
				tail := ulink[head]

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
		state = make([]int, m)
		xcBranch = make([]bool, m)
		cutoff = size

		if debug {
			dump()
		}
	}

	// sitemColors returns a map of secondary items to their currently selected
	// color
	sitemColors := func() *map[string]string {
		// Only one of the secondary items will have it's color value, the
		// others will have -1. Build a map of the colors for each secondary
		// item.
		lcolors := make(map[string]string)

		// Iterate over each value in the current state
		for _, p := range state[0:level] {

			// Cyclically gather the items in the option, beginning at p
			q := p
			for {
				if color[q] > 0 {
					lcolors[name[top[q]]] = colors[color[q]]
				}

				// Advance to the next item
				q++
				if top[q] <= 0 {
					// This is a spacer, so back to the first item
					q = ulink[q]
				}

				if q == p {
					break
				}
			}
		}

		return &lcolors
	}

	showProgress := func() {

		if debug && stats.Verbosity > 0 {
			dump()
		}

		est := 0.0 // estimate of percentage done
		tcum := 1

		lcolors := sitemColors()

		var b strings.Builder
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Current level %d of max %d\n", level, stats.MaxLevel))

		// Iterate over the options
		for _, p := range state[0:level] {

			// Cyclically gather the items in the option, beginning at p
			q := p
			b.WriteString(" ")
			for {
				item := name[top[q]]
				if color, ok := (*lcolors)[item]; ok {
					b.WriteString(fmt.Sprintf(" %v:%s", item, color))
				} else {
					b.WriteString(fmt.Sprintf(" %v", item))
				}

				// Advance to the next item
				q++
				if top[q] <= 0 {
					// This is a spacer, so back to the first item
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

		b.WriteString(fmt.Sprintf("est=%4.4f, %v\n", est, *stats))
		log.Print(b.String())
	}

	// next_item selects the next item to try using these heuristics
	// - Minimum Remaining Values
	// - Sharp Preference
	next_item := func() int {

		i := 0
		theta := -1
		var lambda int
		p := rlink[0]
		for p != 0 {
			if xccOptions.EnableSharpPreference && llen[p] > 1 && name[p][0:1] != "#" {
				lambda = m + llen[p]
			} else {
				lambda = llen[p]
			}
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

		// if p > cutoff {
		// 	// this should not happen
		// 	log.Fatalf("fatal: hiding i=%d with p=%d > cutoff=%d", top[p], p, cutoff)
		// }

		// iterate over the items in this option, skipping p
		q := p + 1
		for q != p {
			x := top[q]
			if x <= 0 {
				// q was a spacer, which is the end of the option, so jump
				// to the first item
				q = ulink[q]
			} else {
				// if color[q] < 0 then it has been purified
				if color[q] >= 0 {
					// remove q from the list for item x
					u, d := ulink[q], dlink[q]
					// if d > cutoff {
					// 	log.Fatalf("fatal in hide: d=%d > cutoff=%d for q=%d", d, cutoff, q)
					// }
					dlink[u], ulink[d] = d, u
					llen[x]--
				}
				// advance to the next item in the option
				q++
			}
		}
	}

	unhide := func(p int) {
		if debug && stats.Verbosity > 1 {
			log.Printf("unhide(p=%d)", p)
		}

		// if p > cutoff {
		// 	log.Fatalf("fatal: unhiding i=%d, q=%d > cutoff=%d", top[p], p, cutoff)
		// }

		// iterate over the items in this option, skipping p, in reverse
		// order
		q := p - 1
		for q != p {
			x := top[q]
			d := dlink[q]
			if x <= 0 {
				// q was a spacer, which is the start of the option, so jump
				// to the last item
				q = d
			} else {
				// if color[q] < 0 then it has been purified
				if color[q] >= 0 {
					if d > cutoff {
						dlink[q], d = x, x
					}
					// restore q back to the list for item x
					u := ulink[q]
					dlink[u], ulink[d] = q, q
					llen[x]++
				}
				// advance to the previous item in the option
				q--
			}
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

	// uncover restores i to the list of items needing to be covered and
	// unhides all of the item's options
	uncover := func(i int) {
		if debug && stats.Verbosity > 1 {
			log.Printf("uncover(i=%d)", i)
		}

		if xccOptions.Minimax {
			// Cutoff items, if necessary
			q := ulink[i]
			// Iterate over items, from the bottom up
			for q > cutoff {
				u := ulink[q]
				// Remove q from the list
				dlink[u], ulink[i] = i, u
				llen[i]--
				q = u
			}
		}

		// restore item i
		l, r := llink[i], rlink[i]
		rlink[l], llink[r] = i, i

		// unhide all of its options, in reverse order
		p := ulink[i]
		for p != i {
			unhide(p)
			p = ulink[p]
		}
	}

	// purify effectively removes all options that have conflicting colors
	// with secondary item p
	purify := func(p int) {
		if debug && stats.Verbosity > 1 {
			log.Printf("purify(p=%d)", p)
		}

		c := color[p] // color of secondary item p
		i := top[p]

		// save color[p] in color[i]; every option with this secondary item
		// will have this same color
		color[i] = c

		// iterate over each option for this secondary item
		q := dlink[i]
		for q != i {
			if color[q] == c {
				// this secondary item has the same color as p, so flag it with
				// value of -1 which indicates this item is already a match, no
				// need to check the color again.
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

		if xccOptions.Minimax {
			// Cutoff items, if necessary
			i := top[p]
			q := ulink[i]
			// Iterate over items, from the bottom up
			for q > cutoff {
				u := ulink[q]
				// Remove q from the list
				dlink[u], ulink[i] = i, u
				llen[i]--
				q = u
			}
		}

		// Iterate over all options for this secondary item p
		c := color[p]
		i := top[p]
		q := ulink[i]
		for q != i {
			if color[q] < 0 {
				// Restore the original color, before purification
				color[q] = c
			} else {
				// dump()
				unhide(q)
			}
			q = ulink[q]
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
		kMax := 0 // level for max p

		// Only one of the secondary items will have it's color value, the
		// others will have -1. Save the color and add it to all the matching
		// secondary items at the end.
		sitemColor := sitemColors()

		// Iterate over the options
		options := make([][]string, 0)
		for i, p := range state[0:level] {
			if p > pMax {
				pMax = p
				kMax = i
			}
			options = append(options, make([]string, 0))

			// Move back to first item in the option
			for top[p-1] > 0 {
				p--
			}

			// Iterate over items in the option
			q := p
			for top[q] > 0 {
				name := name[top[q]]
				if color, ok := (*sitemColor)[name]; ok {
					options[i] = append(options[i], name+":"+color)
				} else {
					options[i] = append(options[i], name)
				}
				q++
			}
		}

		if debug {
			log.Printf("visit(%v)", options)
		}

		if !visit(options) {
			return false
		}

		// For minimax, remove all nodes > cutoff (new value)
		if xccOptions.Minimax {
			// Find spacer at the end of the option for max x_k
			// For minimaxSingle=true find the spacer before the
			// solution, otherwise the spacer after the solution
			pp := pMax
			for top[pp] > 0 {
				if xccOptions.MinimaxSingle {
					pp--
				} else {
					pp++
				}
			}

			// If we have new cutoff value, remove all nodes > cutoff from
			// further consideration
			if pp != cutoff {

				cutoff = pp

				// Iterate over the items of the visited options
				for _, p := range state[0:level] {
					// Cutoff items, if necessary
					x := top[p]
					q := ulink[x]
					// Iterate over items, from the bottom up
					for q > cutoff {
						u := ulink[q]
						// Remove q from the list
						dlink[u], ulink[x] = x, u
						llen[x]--
						q = u
					}
				}
			}

			// Backtrack for each item in state >= lMax
			if xccOptions.MinimaxSingle {
				if debug {
					log.Printf("C2. MinimaxSingle: kMax=%d, pMax=%d, cutoff=%d", kMax, pMax, cutoff)
				}

				for k := level - 1; k >= kMax; k-- {
					i := top[state[k]]
					if debug {
						log.Printf("C2. MinimaxSingle: Backtrack, k=%d, i=%d, Leaving Level %d\n", k, i, level)
					}

					// Uncommit each of the items in this option
					p := state[k] - 1
					for p != state[k] {
						j := top[p]
						if j <= 0 {
							p = dlink[p]
						} else {
							uncommit(p, j)
							p--
						}
					}

					// Uncover item i
					uncover(i)

					level--
				}
			}
		}

		return true
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
		log.Printf("C2. Enter level %d, x[0:l]=%v\n", level, state[0:level])
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
	xcBranch[level] = false
	if xccOptions.Exercise83 && level == 0 {
		if debug && stats.Verbosity > 1 {
			log.Print("Exercise 83: always choose i=1 at level=0")
		}
		i = 1
	} else {
		i = next_item()

		hvForced := name[i][0:1] == "H" || name[i][0:1] == "V"
		xcBranch[level] = !hvForced && level > 0

	}

	if debug {
		if xcBranch[level] {
			log.Printf("C3. Choose connected word")
		} else {
			log.Printf("C3. Choose i=%d (%s)", i, name[i])
		}
	}

	if xcBranch[level] {
		log.Fatal()
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
		log.Printf("C6. Try again, l=%d\n", level)
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

	// Exercise 7.2.2.1-83
	// This code works as expected for Exercise 87.  However, I am unable to
	// reconcile my understanding of this answer to Exercise 83 with the actual
	// description of the exercise.
	// TODO: reconcile this discrepency
	if xccOptions.Exercise83 && level == 0 {

		// x is the first primary item covered
		x := state[0]

		// Find the spacer at the right of this option
		for ; top[x] > 0; x++ {
		}

		// j is the last item in the option
		j = top[x-1]

		if j > n1 && color[x-1] == 0 {
			// j is a secondary item with no color
			// permanently remove from further consideration
			if debug && stats.Verbosity > 1 {
				log.Printf("Exercise 83: covering j=%d\n", j)
			}
			cover(j)
			if debug && stats.Verbosity > 2 {
				dump()
			}
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
