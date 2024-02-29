package cmd

import (
	"fmt"
	"github.com/socialviolation/git-calver/ver"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-calver",
	Short: "CalVer is a git subcommand for managing a calendar versioning tag scheme.",
	Run: func(cmd *cobra.Command, args []string) {
		f, err := ver.NewCalVer(ver.CalVerArgs{Format: "YYYY.MM"})
		if err != nil {
			panic(err)
		}
		v, _ := f.Version(time.Now())
		fmt.Println(v)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-calver.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
