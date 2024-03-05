package cmd

import (
	"fmt"
	pretty "github.com/andanhm/go-prettytime"
	colour "github.com/gookit/color"
	"github.com/socialviolation/git-calver/ver"
	"github.com/spf13/cobra"
)

var noColour bool
var limit int
var listTagCommand = &cobra.Command{
	Use:   "list",
	Short: "Will list all CalVer tags matching the provided format",
	Run: func(cmd *cobra.Command, args []string) {
		f := loadFormat()
		tags, err := ver.ListTags(f, limit)
		CheckIfError(err)

		if len(tags) == 0 {
			fmt.Printf("No tags found.\n")
			return
		}

		if noColour {
			colour.Disable()
		}
		for hash, tags := range tags {
			headline := colour.Green.Sprint(joinTags(tags))
			//if i == 0 {
			//	headline = colour.HEX("#af5fff").Sprint(joinTags(tags))
			//}
			headline += colour.Gray.Sprintf(" - %s", hash)
			ra := fmt.Sprintf("%s - %s", tags[0].Surname(), pretty.Format(tags[0].Time()))
			fmt.Printf("\n%s\n%s\n\t%s", headline, ra, tags[0].Commit.Message)
		}
	},
}

func init() {
	rootCmd.AddCommand(listTagCommand)
	listTagCommand.Flags().BoolVar(&noColour, "no-colour", false, "Disable colour output")
	listTagCommand.Flags().IntVarP(&limit, "limit", "l", 5, "Limit number of results (based on hashes)")
}

func joinTags(tags []ver.CalVerTag) string {
	t := ""
	for i, tag := range tags {
		if i == len(tags)-1 {
			t += fmt.Sprintf("%s", tag.Tag.Name().Short())
			continue
		}
		t += fmt.Sprintf("%s | ", tag.Tag.Name().Short())
	}
	return t
}

/*
idx, err := fuzzyfinder.Find(tags, func(i int) string {
			return tags[i].Short
		}, fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf(`%s`,
				tags[i].FullMessage,
			)
		}))

		if err != nil {
			fmt.Printf("Cancelled.\n")
			return
		}

		fmt.Printf("You selected %d - %s\n", idx, tags[idx].Short)
*/
