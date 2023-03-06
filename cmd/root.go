package cmd

import (
	"fmt"
	"log"
	"os"

	"git.sr.ht/~lbnz/i2/internal/parser"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "i2 [input file]",
	Short: "A verifier that human (mathematicians) can use",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("must specify input file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		file, err := os.ReadFile(args[0])
		if err != nil {
			log.Fatalf("failed to read file: %s\n", err)
		}
		parser.Verify(string(file))
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
