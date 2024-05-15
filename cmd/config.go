package cmd

import (
	"fmt"
	"github.com/socialviolation/git-calver/ver"
	"github.com/spf13/cobra"
)

var formatGetCommand = &cobra.Command{
	Use:   "format",
	Short: "Get format from .gitconfig",
	Run: func(cmd *cobra.Command, args []string) {
		f, a, err := ver.GetRepoFormat()
		CheckIfError(err)

		if a {
			fmt.Println(f.String() + "-[auto-increment]")
			return
		}
		fmt.Println(f.String())
	},
}

var formatSetCommand = &cobra.Command{
	Use:   "set",
	Short: "Set format in .gitconfig",
	Run: func(cmd *cobra.Command, args []string) {
		f, err := ver.NewFormat(format)
		CheckIfError(err)

		err = ver.SetRepoFormat(f)
		CheckIfError(err)

		fmt.Println("format set")
	},
}

func init() {
	rootCmd.AddCommand(formatGetCommand)
	formatGetCommand.AddCommand(formatSetCommand)
	_ = formatSetCommand.MarkFlagRequired("format")
}
