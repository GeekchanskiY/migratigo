package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "migratigo",
	Short: "migratigo is a very fast golang orm",
	Long: `A Fast and Flexible ORM tools built by GeekchanskiY.
			Complete documentation is available at https://github.com/GeekchanskiY/migratigo`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("migratigo")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if _, err := fmt.Fprintln(os.Stderr, err); err != nil {
			panic(err)
		}

		os.Exit(1)
	}
}
