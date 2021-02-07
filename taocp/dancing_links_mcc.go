package taocp

import (
	"fmt"
	"log"
	"math"
	"strings"

	smath "github.com/wallberg/sandbox/math"
)

// MCC implements Algorithm M (7.2.2.1), covering with multiplicities and
// colors via dancing links. The task is to find all subsets of options such
// that:
//
// 1) each primary item j occurs at least u_j times and at most v_j times
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
// visit     -- function called with each discovered solution, returns true
//              if the search should continue
//
func MCC(items []string, multiplicities [][2]int, options [][]string,
	secondary []string,
	stats *ExactCoverStats, visit func(solution [][]string) bool) error {

	var (
		n1       int      // number of primary items
		n2       int      // number of secondary items
		n        int      // total number of items
		nOptions int      // total number of options
		name     []string // name of the item
		llink    []int    // right link of the item
		rlink    []int    // left link of the item
		top      []int
		llen     []int
		ulink    []int
		dlink    []int
		color    []int    // color of a particular item in option
		colors   []string // map of color names, key is the index starting at 1
		level    int      // backtrack level
		state    []int    // search state
		ft       []int    // locations of the "first tweaks" made at a level
		slack    []int    // range of multiplities (v - u)
		bound    []int    // upper limit on number of options for a primary item
		debug    bool     // is debug enabled?
		progress bool     // is progress enabled?
	)

	dump := func() {
		var b strings.Builder
		b.WriteString("\n")

		// Tables
		b.WriteString(fmt.Sprintf("l    :  %d\n", level))
		b.WriteString(fmt.Sprintf("x    :  %v\n", state[0:level]))
		b.WriteString(fmt.Sprintf("ft    :  %v\n", ft[0:level]))
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
		b.WriteString(fmt.Sprintf("slack:  %v\n", slack))
		b.WriteString(fmt.Sprintf("bound:  %v\n", bound))

		// Remaining items
		b.WriteString("items:  ")
		i := 0
		for rlink[i] != 0 {
			i = rlink[i]
			b.WriteString(" " + name[i])
		}
		b.WriteString("\n")

		// Selected options
		for _, p := range state[0:level] {
			b.WriteString(fmt.Sprintf("  option: p=%d (", p))
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

		// Multiplicities
		if len(items) != len(multiplicities) {
			return fmt.Errorf("Number of items, %d, does not match number of multiplicities, %d",
				len(items), len(multiplicities))
		}

		for i, m := range multiplicities {
			u, v := m[0], m[1]
			if u < 0 {
				return fmt.Errorf("Multiplicity i=%d, u=%d, v=%d: u must be >= 0",
					i, u, v)
			}
			if v < 1 {
				return fmt.Errorf("Multiplicity i=%d, u=%d, v=%d: v must be > 0",
					i, u, v)
			}
			if u > v {
				return fmt.Errorf("Multiplicity i=%d, u=%d, v=%d: v must be >= u",
					i, u, v)
			}
		}

		return nil
	}

	initialize := func() {

		n1 = len(items)
		n2 = len(secondary)
		n = n1 + n2
		nOptions = len(options)

		if stats != nil {
			stats.Theta = stats.Delta
			stats.MaxLevel = -1
			if stats.Levels == nil {
				stats.Levels = make([]int, nOptions+1)
			} else {
				for len(stats.Levels) < nOptions+1 {
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

		name[0] = "-"
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
		nOptionItems := 0
		for _, option := range options {
			nOptionItems += len(option)
		}
		size := n + 1 + nOptions + 1 + nOptionItems

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

		// Multiplicities
		slack = make([]int, n1+1)
		bound = make([]int, n1+1)

		for i, m := range multiplicities {
			u, v := m[0], m[1]
			slack[i+1] = v - u
			bound[i+1] = v
		}

		level = 0
		state = make([]int, nOptions)
		ft = make([]int, nOptions)

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
				b.WriteString(fmt.Sprintf("; %d of %d\n", k, llen[i]))
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

	lvisit := func() bool {

		// Only one of the secondary items will have it's color value, the
		// others will have -1. Save the color and add it to all the matching
		// secondary items at the end.
		sitemColor := make(map[string]string)

		// Iterate over the options for this solution
		solution := make([][]string, 0)
		for _, p := range state[0:level] {
			option := make([]string, 0)

			// Determine if this option should be included
			if p <= n1 {
				continue
			}

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
				option = append(option, name)
				q++
			}

			solution = append(solution, option)
		}

		// Add the secondary item colors
		for i, option := range solution {
			for j, item := range option {
				if color, ok := sitemColor[item]; ok {
					solution[i][j] += ":" + color
				}
			}
		}

		return visit(solution)
	}

	// mrv selects the next item to try using the Minimum Remaining
	// Values heuristic of Exercise 166.
	mrv := func() (int, int) {
		i := 0
		theta := math.MaxInt16
		p := rlink[0]
		for p != 0 {
			lambda := smath.MonusInt(llen[p]+1, smath.MonusInt(bound[p], slack[p]))
			if lambda < theta ||
				(lambda == theta && slack[p] < slack[i]) ||
				(lambda == theta && slack[p] == slack[i] && llen[p] > llen[i]) {

				theta = lambda
				i = p
			}
			p = rlink[p]
		}

		return i, theta
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
		for q != p {
			x := top[q]
			u, d := ulink[q], dlink[q]
			if x <= 0 {
				q = d // q was a spacer
			} else {
				if color[q] >= 0 {
					dlink[u], ulink[d] = q, q
					llen[x]++
				}
				q--
			}
		}
	}

	// cover removes i from the list of items needing to be covered removes and
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
				color[q] = -1
			} else {
				hide(q)
			}
			q = dlink[q]
		}
	}

	unpurify := func(p int) {
		if debug && stats.Verbosity > 1 {
			log.Printf("unpurify(p=%d)", p)
		}

		c := color[p]
		i := top[p]
		q := ulink[i]
		for q != i {
			if color[q] < 0 {
				color[q] = c
			} else {
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

	tweak := func(x int, p int, prime bool) {
		if debug && stats.Verbosity > 1 {
			log.Printf("tweak(x=%d, p=%d, prime=%t)", x, p, prime)
		}

		if !prime {
			hide(x)
		}
		d := dlink[x]
		dlink[p] = d
		ulink[d] = p
		llen[p]--

	}

	untweak := func(l int, prime bool) {
		if debug && stats.Verbosity > 1 {
			log.Printf("untweak(l=%d, prime=%t)", l, prime)
		}

		a := ft[l]
		var p int
		if a <= n {
			p = a
		} else {
			p = top[a]
		}
		x := a
		y := p
		z := dlink[p]
		dlink[p] = x
		k := 0
		for x != z {
			ulink[x] = y
			k++
			if !prime {
				unhide(x)
			}
			y = x
			x = dlink[x]
		}
		ulink[z] = y
		llen[p] += k
		if prime {
			uncover(p)
		}
	}

	// M1 [Initialize.]
	if stats != nil && stats.Debug {
		log.Printf("M1. Initialize")
	}

	if err := validate(); err != nil {
		return err
	}
	initialize()

	var (
		i     int
		j     int
		p     int
		q     int
		theta int
	)

	if progress {
		showProgress()
	}

M2:
	// M2. [Enter level l.]
	if debug {
		log.Printf("M2. l=%d, x[0:l]=%v\n", level, state[0:level])
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
			log.Println("M2. Visit the solution")
		}
		if stats != nil {
			stats.Solutions++
		}
		resume := lvisit()
		if !resume {
			if debug {
				log.Println("M2. Halting the search")
			}
			if progress {
				showProgress()
			}
			return nil
		}
		goto M9
	}

	// M3. [Choose i.]
	i, theta = mrv()

	if debug {
		log.Printf("M3. Choose i=%d (%s), theta=%d\n", i, name[i], theta)
	}

	if theta == 0 {
		goto M9
	}

	// M4. [Prepare to branch on i.]
	if debug {
		log.Printf("M4. Prepare to branch on i=%d (%s)\n", i, name[i])
	}
	state[level] = dlink[i]
	bound[i]--
	if bound[i] == 0 {
		cover(i)
	}
	if bound[i] != 0 || slack[i] != 0 {
		ft[level] = state[level]
	}

M5:
	// M5. [Possibly tweak x_l.]
	if debug {
		log.Printf("M5. Possibly tweak l=%d, x[0:l]=%v\n", level, state[0:level])
	}

	if bound[i] == 0 && slack[i] == 0 {
		if state[level] != i {
			goto M6
		}
		goto M8

	} else if llen[i] <= bound[i]-slack[i] {
		// list i is too short
		goto M8

	} else if state[level] != i {
		tweak(state[level], i, bound[i] == 0)

	} else if bound[i] != 0 {
		p = llink[i]
		q = rlink[i]
		rlink[p] = q
		llink[q] = p
	}

M6:
	// M6. [Try x_l.]
	if debug {
		log.Printf("M6. Try l=%d, x[0:l]=%v\n", level, state[0:level])
	}
	if state[level] != i {
		p = state[level] + 1
		// Cover or partially cover the items != i in the option that contains
		// x_l
		for p != state[level] {
			j = top[p]
			if j <= 0 {
				p = ulink[p]
			} else if j <= n1 {
				bound[j]--
				p++
				if bound[j] == 0 {
					cover(j)
				}
			} else {
				commit(p, j)
				p++
			}
		}
	}
	level++
	goto M2

M7:
	// M7. [Try again.]
	if debug {
		log.Println("M7. Try again")
	}

	if stats != nil {
		stats.Nodes++
	}

	// Uncommit each of the items in this option
	p = state[level] - 1

	// Uncover the items != i in the option that contains x_l, using the
	// reverse order
	for p != state[level] {
		j = top[p]
		if j <= 0 {
			p = dlink[p]
		} else if j <= n1 {
			bound[j]++
			p--
			if bound[j] == 1 {
				uncover(j)
			}
		} else {
			uncommit(p, j)
			p--
		}
	}
	state[level] = dlink[state[level]]
	goto M5

M8:
	// M8. [Restore i.]
	if debug {
		log.Println("M8. Restore i")
	}
	if bound[i] == 0 && slack[i] == 0 {
		uncover(i)
	} else {
		untweak(level, bound[i] == 0)
	}
	bound[i]++

M9:
	// M9. [Leave level l.]
	if debug {
		log.Printf("M9. Leaving level %d\n", level)
	}
	if level == 0 {
		if progress {
			showProgress()
		}
		return nil
	}
	level--
	if state[level] <= n {
		// Reactivate i
		i = state[level]
		p = llink[i]
		q = rlink[i]
		rlink[p] = i
		llink[q] = i
		goto M8
	}
	i = top[state[level]]
	goto M7
}
