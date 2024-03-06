package cmd

import (
	"fmt"
	colour "github.com/gookit/color"
	"github.com/socialviolation/git-calver/ver"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	noColour  bool
	limit     int
	changelog bool
)

var nextTagCommand = &cobra.Command{
	Use:   "latest",
	Short: "Output what the next calver tag will be",
	Run: func(cmd *cobra.Command, args []string) {
		f := loadFormat()
		tag, err := ver.LatestTag(f, changelog)
		CheckIfError(err)

		fmt.Println(tag)
	},
}

var latestTagCommand = &cobra.Command{
	Use:   "latest",
	Short: "Get latest tag matching the provided format",
	Run: func(cmd *cobra.Command, args []string) {
		f := loadFormat()
		tag, err := ver.LatestTag(f, changelog)
		CheckIfError(err)

		if tag == nil {
			fmt.Printf("No tag found.\n")
			return
		}

		tag.Print(os.Stdout, noColour)
	},
}

var listTagCommand = &cobra.Command{
	Use:   "list",
	Short: "Will list all CalVer tags matching the provided format",
	Run: func(cmd *cobra.Command, args []string) {
		f := loadFormat()
		tags, err := ver.ListTags(f, limit, changelog)
		CheckIfError(err)

		if len(tags) == 0 {
			fmt.Printf("No tags found.\n")
			return
		}

		for _, tag := range tags {
			tag.Print(os.Stdout, noColour)
		}
	},
}

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "tag",
	Run: func(cmd *cobra.Command, args []string) {
		cf := loadFormat()
		f, err := ver.NewCalVer(
			ver.CalVerArgs{
				Format:   cf,
				Micro:    &micro,
				Minor:    &minor,
				Modifier: modifier,
			})
		CheckIfError(err)
		v, _ := f.Version(time.Now())
		fmt.Println(v)
		err = ver.TagNext(ver.TagArgs{
			Hash: hash,
			Push: push,
			CV:   f,
		})
		if err != nil {
			colour.Red.Println("error: %s", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listTagCommand)
	listTagCommand.Flags().BoolVar(&noColour, "no-colour", false, "Disable colour output")
	listTagCommand.Flags().BoolVar(&changelog, "changeLog", true, "Include changelog")
	listTagCommand.Flags().IntVarP(&limit, "limit", "l", 5, "Limit number of results (based on hashes)")

	rootCmd.AddCommand(latestTagCommand)
	latestTagCommand.Flags().BoolVar(&noColour, "no-colour", false, "Disable colour output")
	latestTagCommand.Flags().BoolVar(&changelog, "changeLog", true, "Include changelog")

	rootCmd.AddCommand(tagCmd)
	tagCmd.Flags().BoolVarP(&push, "push", "p", false, "Push tag after create")
	tagCmd.Flags().StringVarP(&hash, "hash", "h", "", "Override with Hash")

	rootCmd.AddCommand(nextTagCommand)
}
