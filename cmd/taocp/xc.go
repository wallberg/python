package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"

	"github.com/wallberg/sandbox/taocp"
	"gopkg.in/yaml.v2"
)

// initialize this command by adding it to the parser
func init() {
	var command xcCommand

	_, err := parser.AddCommand("xc",
		"Exact Cover w/ Colors (XCC)",
		`Solve Exact Cover w/ Colors (XCC) problems using taocp.ExactCoverColors
Uses YAML for input and output`,
		&command,
	)
	if err != nil {
		log.Fatalf("Error adding xc command: %v", err)
	}
}

type xcCommand struct {
	Input     string `short:"i" long:"input" description:"Input YAML" default:"-"`
	Output    string `short:"o" long:"output" description:"Output YAML" default:"-"`
	Verbosity int    `short:"v" long:"verbosity" description:"Verbosity level" default:"1"`
	Delta     int    `short:"d" long:"delta" description:"Display progress ~Delta nodes (Verbosity > 0)" default:"10000000"`
}

func (command xcCommand) Execute(args []string) error {
	var err error

	// Open input file for reading
	var input *os.File
	if command.Input == "-" {
		input = os.Stdin
	} else {
		if input, err = os.Open(command.Input); err != nil {
			return err
		}
	}
	defer input.Close()

	// Read the input
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, input); err != nil {
		return err
	}

	// Deserialize from YAML
	var xcYaml taocp.ExactCoverYaml
	err = yaml.Unmarshal(buf.Bytes(), &xcYaml)
	if err != nil {
		return err
	}
	options := make([][]string, len(xcYaml.Options))
	for i, option := range xcYaml.Options {
		options[i] = strings.Split(option, " ")
	}

	// Open output file for writing
	var output *os.File
	if command.Output == "-" {
		output = os.Stdout
	} else {
		if output, err = os.Create(command.Output); err != nil {
			return err
		}
	}
	defer output.Close()

	// Solve
	stats := &taocp.ExactCoverStats{
		Debug:     command.Verbosity > 1,
		Progress:  command.Verbosity > 0,
		Verbosity: command.Verbosity - 2,
		Delta:     command.Delta,
	}
	output.WriteString("solutions:\n")
	err = taocp.XCC(xcYaml.Items, options, xcYaml.SItems, stats, false, false,
		func(solution [][]string) bool {
			output.WriteString("  -\n")
			for _, option := range solution {
				output.WriteString("    - \"")
				output.WriteString(strings.Join(option, " "))
				output.WriteString("\"\n")
			}
			return true
		})
	if err != nil {
		return err
	}

	log.Println("Stats:", *stats)

	return nil
}
