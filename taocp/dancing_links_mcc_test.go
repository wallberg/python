package taocp

import (
	"log"
	"reflect"
	"testing"
)

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

		// {
		// 	[]string{"#1", "#2", "00", "01", "10", "11"},
		// 	[][2]int{{2, 2}, {0, 1}, {1, 1}, {1, 1}, {1, 1}, {1, 1}},
		// 	[][]string{
		// 		{"#1", "00"},
		// 		{"#1", "01"},
		// 		{"#1", "10"},
		// 		{"#1", "11"},
		// 		{"#2", "00", "10"},
		// 		{"#2", "10", "11"},
		// 		{"#2", "01", "11"},
		// 		{"#2", "00", "10"},
		// 	},
		// 	[]string{},
		// 	[][][]string{
		// 		{{"#1", "01"}, {"#1", "00"}, {"#2", "10", "11"}},
		// 		{{"#1", "01"}, {"#1", "00"}, {"#2", "10", "11"}},
		// 		{{"#1", "01"}, {"#1", "11"}, {"#2", "00", "10"}},
		// 		{{"#1", "01"}, {"#1", "11"}, {"#2", "00", "10"}},
		// 	},
		// },
		// {
		// 	[]string{"#1", "#2", "00", "01", "10", "11"},
		// 	[][2]int{{2, 2}, {0, 2}, {1, 1}, {1, 1}, {1, 1}, {1, 1}},
		// 	[][]string{
		// 		{"#1", "00"},
		// 		{"#1", "01"},
		// 		{"#1", "10"},
		// 		{"#1", "11"},
		// 		{"#2", "00", "10"},
		// 		{"#2", "10", "11"},
		// 		{"#2", "01", "11"},
		// 		{"#2", "00", "10"},
		// 	},
		// 	[]string{},
		// 	[][][]string{
		// 		{{"#1", "01"}, {"#1", "00"}, {"#2", "10", "11"}},
		// 		{{"#1", "01"}, {"#1", "00"}, {"#2", "10", "11"}},
		// 		{{"#1", "01"}, {"#1", "11"}, {"#2", "00", "10"}},
		// 		{{"#1", "01"}, {"#1", "11"}, {"#2", "00", "10"}},
		// 	},
		// },
	}

	for _, c := range cases {
		got := make([][][]string, 0)
		stats := &ExactCoverStats{Progress: true, Delta: 0, Debug: true, Verbosity: 2}
		err := MCC(c.items, c.multiplicities, c.options, c.secondary, stats,
			func(solution [][]string) bool {
				log.Print(solution)
				got = append(got, solution)
				return true
			})

		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(got, c.solutions) {
			t.Errorf("Got solutions %v; want %v", got, c.solutions)
		}

		// t.Error("True")
	}
}
