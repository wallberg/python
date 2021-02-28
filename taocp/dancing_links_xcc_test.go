package taocp

import (
	"log"
	"reflect"
	"testing"
)

var (
	// Toy XCC example 7.2.2.1-49
	xccItems = []string{"p", "q", "r"}

	xccSItems = []string{"x", "y"}

	xccOptions = [][]string{
		{"p", "q", "x", "y:A"},
		{"p", "r", "x:A", "y"},
		{"p", "x:B"},
		{"q", "x:A"},
		{"r", "y:B"},
	}

	xccExpected = [][]string{
		{"q", "x:A"},
		{"p", "r", "x:A", "y"},
	}
)

func TestXCC(t *testing.T) {

	var (
		count int
		stats *ExactCoverStats
	)

	count = 0
	stats = &ExactCoverStats{
		// Progress:  true,
		// Delta:     0,
		// Debug:     true,
		// Verbosity: 2,
	}

	XCC(xcItems, xcOptions, []string{}, stats, false, false,
		func(solution [][]string) bool {
			if !reflect.DeepEqual(solution, xcExpected) {
				t.Errorf("Expected %v; got %v", xcExpected, solution)
			}
			count++
			return true
		})

	if count != 1 {
		t.Errorf("Expected 1 solution; got %d", count)
	}

	if stats.Solutions != 1 {
		t.Errorf("Expected 1 stats.Solution; got %d", stats.Solutions)
	}

	count = 0
	stats = &ExactCoverStats{
		// Progress:  true,
		// Delta:     0,
		// Debug:     true,
		// Verbosity: 2,
	}
	XCC(xccItems, xccOptions, xccSItems, stats, false, false,
		func(solution [][]string) bool {
			if !reflect.DeepEqual(solution, xccExpected) {
				t.Errorf("Expected %v; got %v", xccExpected, solution)
			}
			count++
			return true
		})

	if count != 1 {
		t.Errorf("Expected 1 solution; got %d", count)
	}

	if stats.Solutions != 1 {
		t.Errorf("Expected 1 stats.Solution; got %d", stats.Solutions)
	}
}

var (
	cards1 = [9][3][3]int{
		{{1, 0, 0}, {0, 2, 0}, {8, 0, 3}},
		{{2, 0, 0}, {0, 3, 0}, {1, 0, 4}},
		{{3, 0, 0}, {0, 4, 0}, {1, 0, 5}},
		{{4, 0, 0}, {0, 5, 0}, {2, 0, 6}},
		{{5, 0, 0}, {0, 6, 0}, {4, 0, 7}},
		{{6, 0, 0}, {0, 7, 0}, {4, 0, 8}},
		{{7, 0, 0}, {0, 8, 0}, {5, 0, 9}},
		{{8, 0, 0}, {0, 9, 0}, {7, 0, 1}},
		{{9, 0, 0}, {0, 1, 0}, {7, 0, 2}},
	}

	cards1Expected = [9]int{1, 9, 2, 4, 3, 5, 7, 6, 8}

	grid1Expected = [9][9]int{
		{1, 4, 5, 9, 8, 3, 2, 7, 6},
		{6, 2, 7, 5, 1, 4, 9, 3, 8},
		{8, 9, 3, 7, 6, 2, 1, 5, 4},
		{4, 7, 8, 3, 2, 6, 5, 1, 9},
		{9, 5, 1, 8, 4, 7, 3, 6, 2},
		{2, 3, 6, 1, 9, 5, 4, 8, 7},
		{7, 1, 2, 6, 5, 9, 8, 4, 3},
		{3, 8, 4, 2, 7, 1, 6, 9, 5},
		{5, 6, 9, 4, 3, 8, 7, 2, 1},
	}

	cards2 = [9][3][3]int{
		{{1, 0, 0}, {0, 2, 0}, {9, 0, 3}},
		{{2, 0, 0}, {0, 3, 0}, {9, 0, 4}},
		{{3, 0, 0}, {0, 4, 0}, {8, 0, 5}},
		{{4, 0, 0}, {0, 5, 0}, {1, 0, 6}},
		{{5, 0, 0}, {0, 6, 0}, {3, 0, 7}},
		{{6, 0, 0}, {0, 7, 0}, {5, 0, 8}},
		{{7, 0, 0}, {0, 8, 0}, {2, 0, 9}},
		{{8, 0, 0}, {0, 9, 0}, {6, 0, 1}},
		{{9, 0, 0}, {0, 1, 0}, {4, 0, 2}},
	}

	cards2Expected = [9]int{1, 4, 9, 5, 2, 3, 7, 8, 6}

	grid2Expected = [9][9]int{
		{1, 5, 6, 4, 2, 7, 9, 8, 3},
		{4, 2, 8, 3, 5, 9, 7, 1, 6},
		{9, 7, 3, 1, 8, 6, 4, 5, 2},
		{5, 9, 4, 2, 1, 8, 3, 6, 7},
		{8, 6, 2, 7, 3, 5, 1, 4, 9},
		{3, 1, 7, 9, 6, 4, 8, 2, 5},
		{7, 3, 5, 8, 4, 2, 6, 9, 1},
		{6, 8, 1, 5, 9, 3, 2, 7, 4},
		{2, 4, 9, 6, 7, 1, 5, 3, 8},
	}
)

func TestSudokuCards(t *testing.T) {

	cases := []struct {
		cards         [9][3][3]int
		cardsExpected [9]int
		gridExpected  [9][9]int
	}{
		{cards1, cards1Expected, grid1Expected},
		{cards2, cards2Expected, grid2Expected},
	}

	for _, c := range cases {
		count := 0
		stats := new(ExactCoverStats)

		SudokuCards(c.cards, stats,
			func(solution [9]int, grid [9][9]int) bool {
				if !reflect.DeepEqual(solution, c.cardsExpected) {
					t.Errorf("Expected card ordering %v; got %v", c.cardsExpected, solution)
				}
				if !reflect.DeepEqual(grid, c.gridExpected) {
					t.Errorf("Expected grid %v; got %v", c.gridExpected, grid)
				}
				count++
				return true
			})

		if count != 1 {
			t.Errorf("Expected 1 solution; got %d", count)
		}
	}
}

func BenchmarkSudokuCards(b *testing.B) {
	cases := []struct {
		name  string
		cards [9][3][3]int
	}{
		{"cards1", cards1},
		{"cards2", cards2},
	}

	for _, c := range cases {

		b.Run(c.name, func(b *testing.B) {
			for repeat := 0; repeat < b.N; repeat++ {
				SudokuCards(cards2, nil, func([9]int, [9][9]int) bool { return true })
			}
		})
	}
}

func TestXCCminimax(t *testing.T) {

	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	cases := []struct {
		items     []string
		options   [][]string
		secondary []string
		single    bool
		solutions [][][]string
	}{
		{
			[]string{"a"},
			[][]string{
				{"a", "x"},
				{"a", "y"},
				{"a", "z"},
			},
			[]string{"x", "y", "z"},
			false,
			[][][]string{
				{{"a", "x"}},
			},
		},
		{
			[]string{"a", "b"},
			[][]string{
				{"a", "x"},
				{"a", "y"},
				{"a", "z"},
				{"b", "y"},
			},
			[]string{"x", "y", "z"},
			false,
			[][][]string{
				{{"b", "y"}, {"a", "x"}},
				{{"b", "y"}, {"a", "z"}},
			},
		},
		{
			[]string{"a", "b"},
			[][]string{
				{"a", "x"},
				{"a", "y"},
				{"a", "z"},
				{"b", "y"},
			},
			[]string{"x", "y", "z"},
			true,
			[][][]string{
				{{"b", "y"}, {"a", "x"}},
			},
		},
		{
			[]string{"a", "b"},
			[][]string{
				{"a", "x"},
				{"a", "y"},
				{"b", "y"},
				{"b", "x"},
			},
			[]string{"x", "y", "z"},
			true,
			[][][]string{
				{{"a", "x"}, {"b", "y"}},
			},
		},
		{
			[]string{"a", "b", "c", "d"},
			[][]string{
				{"a", "b", "x"},
				{"a", "b", "y:1"},
				{"b", "c", "y"},
				{"b", "c", "x"},
				{"a"},
				{"b"},
				{"c", "y:2"},
				{"c", "y:3"},
				{"c", "d", "z"},
				{"d", "y:3"},
				{"c", "d", "y"},
				{"c", "d", "x"},
			},
			[]string{"x", "y", "z"},
			false,
			[][][]string{
				{{"a", "b", "x"}, {"c", "d", "z"}},
				{{"a", "b", "y:1"}, {"c", "d", "z"}},
				{{"a"}, {"c", "d", "z"}, {"b"}},
			},
		},
		{
			[]string{"a", "b", "c", "d"},
			[][]string{
				{"a", "b", "y:1"},
				{"b", "c", "y"},
				{"b", "c", "x"},
				{"a", "b", "x"},
				{"a"},
				{"b"},
				{"c", "y:2"},
				{"c", "y:3"},
				{"c", "d", "z"},
				{"d", "y:3"},
				{"c", "d", "y"},
				{"c", "d", "x"},
			},
			[]string{"x", "y", "z"},
			false,
			[][][]string{
				{{"a", "b", "x"}, {"c", "d", "z"}},
				{{"a", "b", "y:1"}, {"c", "d", "z"}},
				{{"a"}, {"c", "d", "z"}, {"b"}},
			},
		},
		{
			[]string{"a", "b", "c", "d"},
			[][]string{
				{"a", "b", "x"},
				{"a", "b", "y:1"},
				{"b", "c", "y"},
				{"b", "c", "x"},
				{"a"},
				{"b"},
				{"c", "y:2"},
				{"d", "y:3"},
				{"c", "d", "z"},
				{"c", "d", "y"},
				{"c", "d", "x"},
			},
			[]string{"x", "y", "z"},
			true,
			[][][]string{
				{{"a", "b", "x"}, {"c", "d", "z"}},
			},
		},
	}

	for i, c := range cases {
		got := make([][][]string, 0)
		stats := &ExactCoverStats{
			// Progress:  true,
			// Delta:     0,
			// Debug:     true,
			// Verbosity: 2,
		}
		err := XCC(c.items, c.options, c.secondary, stats, true, c.single,
			func(solution [][]string) bool {
				got = append(got, solution)
				return true
			})

		if err != nil {
			t.Error(err)
		}

		sortSolutions(got)
		sortSolutions(c.solutions)

		if !reflect.DeepEqual(got, c.solutions) {
			t.Errorf("For case #%d, got solutions %v; want %v", i, got, c.solutions)
		}
	}
}
