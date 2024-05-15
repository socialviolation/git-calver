package cmd

import (
	"fmt"
	colour "github.com/gookit/color"
	"github.com/socialviolation/git-calver/ver"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	dryRun   bool
	format   string
	minor    uint
	micro    uint
	modifier string

	hash string
	push bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-calver",
	Short: "CalVer is a git subcommand for managing a calendar versioning tag scheme.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !autoIncrementFlag && !cmd.Flags().Changed("auto-increment") {
			autoIncrement = false
		}
		if autoIncrementFlag {
			autoIncrement = true
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		f := latestCalVer()
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
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry run")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "", "format of calver (YYYY.0M.0D)")
	rootCmd.PersistentFlags().StringVar(&modifier, "modifier", "", "Modifer (eg. DEV, RC, etc)")
	rootCmd.PersistentFlags().UintVar(&minor, "minor", 0, "Minor Version")
	rootCmd.PersistentFlags().UintVar(&micro, "micro", 0, "Micro Version")
}

func latestCalVer() *ver.CalVer {
	cf := loadFormat()
	f, err := ver.NewCalVer(
		ver.CalVerArgs{
			Format:        cf,
			Micro:         &micro,
			Minor:         &minor,
			Modifier:      modifier,
			AutoIncrement: autoIncrement,
		})
	CheckIfError(err)
	return f
}

func nextCalVerArgs() *ver.CalVer {
	f := loadFormat()
	cv, err := ver.NextCalVer(
		ver.CalVerArgs{
			Format:        f,
			Micro:         &micro,
			Minor:         &minor,
			Modifier:      modifier,
			AutoIncrement: autoIncrement,
		})
	CheckIfError(err)
	return cv
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}
	colour.Red.Printf("error: %s\n", err)
	os.Exit(1)
}

func loadFormat() *ver.Format {
	f, source, err := getFormat()
	if err != nil {
		if err.Error() == "format not set" {
			colour.Red.Println("format not set, please set with --format or CALVER environment variable or git config")
			os.Exit(1)
		}
		colour.Red.Printf("loading from %s error: %s\n", source, err)
		os.Exit(1)
	}
	return f
}

func getFormat() (*ver.Format, string, error) {
	if format != "" {
		f, err := ver.NewFormat(format)
		if err != nil {
			return nil, "argument", err
		}
		if strings.HasSuffix(format, "-AUTO") {
			autoIncrement = true
		}

		return f, "argument", nil
	}

	envVar := os.Getenv("CALVER")
	if envVar != "" {
		f, err := ver.NewFormat(envVar)
		if err != nil {
			return nil, "environment", err
		}
		if strings.HasSuffix(envVar, "-AUTO") {
			autoIncrement = true
		}

		return f, "environment", nil
	}

	gitConf, a, err := ver.GetRepoFormat()
	if err != nil {
		if err.Error() == "[calver] not set" {
			return nil, "gitconfig", fmt.Errorf("format not set")
		}
		return nil, "gitconfig", err
	}
	if a {
		autoIncrement = true
	}

	return gitConf, "gitconfig", nil
}
