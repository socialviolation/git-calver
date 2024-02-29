package cmd

import (
	"fmt"
	"github.com/socialviolation/git-calver/git"
	"github.com/socialviolation/git-calver/ver"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	format   string
	minor    uint
	micro    uint
	modifier string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-calver",
	Short: "CalVer is a git subcommand for managing a calendar versioning tag scheme.",
	Run: func(cmd *cobra.Command, args []string) {
		f, err := ver.NewCalVer(ver.CalVerArgs{Format: "YYYY.MM"})
		CheckIfError(err)
		v, _ := f.Version(time.Now())
		fmt.Println(v)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	CheckIfError(err)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "", "format of calver")
	rootCmd.PersistentFlags().StringVar(&modifier, "modifier", "f", "format of calver (YYYY.0M.0D)")
	rootCmd.PersistentFlags().UintVar(&minor, "minor", 0, "Minor Version")
	rootCmd.PersistentFlags().UintVar(&micro, "micro", 0, "Micro Version")
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func loadFormat() *ver.Format {
	f, source, err := getFormat()
	if err != nil {
		fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("loading from %s error: %s", source, err))
		os.Exit(1)
	}
	fmt.Printf("loaded %s from %s\n", f.String(), source)
	return f
}

func getFormat() (*ver.Format, string, error) {
	if format != "" {
		f, err := ver.NewFormat(format)
		if err != nil {
			return nil, "argument", err
		}

		return f, "argument", nil
	}

	envVar := os.Getenv("CALVAR")
	if envVar != "" {
		f, err := ver.NewFormat(envVar)
		if err != nil {
			return nil, "environment", err
		}

		return f, "environment", nil
	}

	gitConf, err := git.GetFormat()
	if err != nil {
		return nil, "gitconfig", err
	}

	return gitConf, "gitconfig", nil
}
