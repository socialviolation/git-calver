package cmd

import (
	"fmt"
	"github.com/socialviolation/git-calver/ver"
	"github.com/spf13/cobra"
)

var listTagCommand = &cobra.Command{
	Use:   "list",
	Short: "Will list all CalVer tags matching the provided format",
	Run: func(cmd *cobra.Command, args []string) {
		f := loadFormat()
		tags, err := ver.ListTags(f)
		CheckIfError(err)

		if len(tags) == 0 {
			fmt.Printf("No tags found.\n")
			return
		}

		for _, tag := range tags {
			fmt.Printf("%s (%s)\n", tag.Short, tag.Time.Format("15:04 2006-01-02"))
		}
	},
}

func init() {
	rootCmd.AddCommand(listTagCommand)
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
