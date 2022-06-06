package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/formatter"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/validation"
)

type validateCommand struct {
	outputFormatter *formatter.Output
}

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your Keel schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := &validateCommand{
			outputFormatter: formatter.New(os.Stdout),
		}

		switch outputFormat {
		case string(formatter.FormatJSON):
			c.outputFormatter.SetOutput(formatter.FormatJSON, os.Stdout)
		default:
			c.outputFormatter.SetOutput(formatter.FormatText, os.Stdout)
		}

		schema := schema.Schema{}
		var err error

		switch {
		case inputFile != "":
			_, err = schema.MakeFromFile(inputFile)
		default:
			_, err = schema.MakeFromDirectory(inputDir)
		}

		if err != nil {
			errs, ok := err.(validation.ValidationErrors)
			if ok {
				return c.outputFormatter.Write(errs.ToAnnotatedSchema())
			} else {
				return fmt.Errorf("error making schema: %v", err)
			}
		}

		c.outputFormatter.Write([]byte(color.New(color.FgGreen).Sprint("Validation OK\n")))

		return nil
	},
}

var inputDir string
var inputFile string
var outputFormat string

func init() {
	rootCmd.AddCommand(validateCmd)
	defaultDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("os.Getwd() errored: %v", err))
	}
	validateCmd.Flags().StringVarP(&inputDir, "dir", "d", defaultDir, "input directory to validate")
	validateCmd.Flags().StringVarP(&inputFile, "file", "f", "", "schema file to validate")
	validateCmd.Flags().StringVarP(&outputFormat, "output", "o", "console", "output format (console, json)")
}
