package taocp

import (
	"log"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func sortSolutions(solutions [][][]string) {
	for _, solution := range solutions {
		for _, option := range solution {
			// Sort the items in the option
			sort.Strings(option)
		}

		// Sort the options in the solution
		sort.Slice(solution, func(i, j int) bool {
			return strings.Join(solution[i], " ") < strings.Join(solution[j], " ")
		})
	}

	// Sort the solutions
	sort.Slice(solutions, func(i, j int) bool {
		k := 0
		for ; k < len(solutions[i]) && k < len(solutions[j]); k++ {
			iOption := strings.Join(solutions[i][k], " ")
			jOption := strings.Join(solutions[j][k], " ")

			if iOption < jOption {
				return true
			} else if iOption > jOption {
				return false
			}
		}

		if k < len(solutions[j]) {
			// We ran out of i options first
			return true
		}

		// We ran out of j options first or at the same time
		return false

	})
}

func TestMCC(t *testing.T) {

	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	cases := []struct {
		items          []string
		multiplicities [][2]int
		options        [][]string
		secondary      []string
		solutions      [][][]string
	}{
		{
			[]string{"a", "b"},
			[][2]int{{0, 1}, {3, 3}},
			[][]string{
				{"a", "b"},
				{"a"},
				{"b"},
			},
			[]string{},
			[][][]string{},
		},

		{
			[]string{"a", "b"},
			[][2]int{{1, 1}, {2, 3}},
			[][]string{
				{"a", "b"},
				{"a"},
				{"b"},
			},
			[]string{},
			[][][]string{
				{{"a", "b"}, {"b"}},
			},
		},

		{
			[]string{"a", "b"},
			[][2]int{{1, 1}, {1, 1}},
			[][]string{
				{"a", "b"},
				{"a"},
				{"b"},
			},
			[]string{},
			[][][]string{
				{{"a", "b"}},
				{{"a"}, {"b"}},
			},
		},

		{
			[]string{"a", "b"},
			[][2]int{{0, 1}, {1, 1}},
			[][]string{
				{"a", "b"},
				{"a"},
				{"b"},
			},
			[]string{},
			[][][]string{
				{{"a", "b"}},
				{{"b"}, {"a"}},
				{{"b"}},
			},
		},

		{
			[]string{"a", "b"},
			[][2]int{{0, 1}, {1, 2}},
			[][]string{
				{"a", "b"},
				{"a"},
				{"b"},
			},
			[]string{},
			[][][]string{
				{{"a", "b"}, {"b"}},
				{{"a", "b"}},
				{{"b"}, {"a"}},
				{{"b"}},
			},
		},

		{
			[]string{"a", "b"},
			[][2]int{{0, 2}, {0, 2}},
			[][]string{
				{"a", "b"},
				{"a"},
				{"b"},
			},
			[]string{},
			[][][]string{
				{{"a", "b"}, {"a"}, {"b"}},
				{{"a", "b"}, {"a"}},
				{{"a", "b"}, {"b"}},
				{{"a", "b"}},
				{{"a"}, {"b"}},
				{{"a"}},
				{{"b"}},
				{},
			},
		},

		{
			[]string{"#1", "#2", "00", "01", "10", "11"},
			[][2]int{{2, 2}, {0, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}},
			[][]string{
				{"#1", "00"},
				{"#1", "01"},
				{"#1", "10"},
				{"#1", "11"},
				{"#2", "00", "01"},
				{"#2", "10", "11"},
				{"#2", "01", "11"},
				{"#2", "00", "10"},
			},
			[]string{},
			[][][]string{
				{{"#1", "00"}, {"#1", "01"}, {"#2", "10", "11"}},
				{{"#1", "00"}, {"#1", "10"}, {"#2", "01", "11"}},
				{{"#1", "01"}, {"#1", "11"}, {"#2", "00", "10"}},
				{{"#1", "10"}, {"#1", "11"}, {"#2", "00", "01"}},
			},
		},

		{
			[]string{"#1", "#2", "00", "01", "10", "11"},
			[][2]int{{2, 3}, {0, 2}, {1, 1}, {1, 1}, {1, 1}, {1, 1}},
			[][]string{
				{"#1", "00"},
				{"#1", "01"},
				{"#1", "10"},
				{"#1", "11"},
				{"#2", "00", "10"},
				{"#2", "10", "11"},
				{"#2", "01", "11"},
				{"#2", "00", "01"},
			},
			[]string{},
			[][][]string{
				{{"#1", "00"}, {"#1", "01"}, {"#2", "10", "11"}},
				{{"#1", "00"}, {"#1", "10"}, {"#2", "01", "11"}},
				{{"#1", "01"}, {"#1", "11"}, {"#2", "00", "10"}},
				{{"#1", "10"}, {"#1", "11"}, {"#2", "00", "01"}},
			},
		},

		{
			[]string{"#1", "#2", "00", "01", "10", "11"},
			[][2]int{{0, 4}, {0, 2}, {1, 1}, {1, 1}, {1, 1}, {1, 1}},
			[][]string{
				{"#1", "00"},
				{"#1", "01"},
				{"#1", "10"},
				{"#1", "11"},
				{"#2", "00", "10"},
				{"#2", "10", "11"},
				{"#2", "01", "11"},
				{"#2", "00", "01"},
			},
			[]string{},
			[][][]string{
				{{"#1", "00"}, {"#1", "01"}, {"#1", "10"}, {"#1", "11"}},
				{{"#1", "00"}, {"#1", "01"}, {"#2", "10", "11"}},
				{{"#1", "00"}, {"#1", "10"}, {"#2", "01", "11"}},
				{{"#1", "01"}, {"#1", "11"}, {"#2", "00", "10"}},
				{{"#1", "10"}, {"#1", "11"}, {"#2", "00", "01"}},
				{{"#2", "00", "01"}, {"#2", "10", "11"}},
				{{"#2", "00", "10"}, {"#2", "01", "11"}},
			},
		},
	}

	for _, c := range cases {
		got := make([][][]string, 0)
		stats := &ExactCoverStats{
			Progress:  false,
			Delta:     0,
			Debug:     false,
			Verbosity: 2,
		}
		err := MCC(c.items, c.multiplicities, c.options, c.secondary, stats,
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
			t.Errorf("Got solutions %v; want %v", got, c.solutions)
		}
	}
}
