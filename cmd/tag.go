package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	colour "github.com/gookit/color"
	"github.com/socialviolation/git-calver/ver"
	"github.com/spf13/cobra"
)

var (
	noColour      bool
	limit         int
	changelog     bool
	autoIncrement bool
	lean          bool
)

var latestTagCmd = &cobra.Command{
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

var nextTagCommand = &cobra.Command{
	Use:   "next",
	Short: "Output what the next calver tag will be",
	Run: func(cmd *cobra.Command, args []string) {
		f := loadFormat()
		verifiedHash, err := ver.VerifyHash(hash)
		CheckIfError(err)
		tag := f.Version(time.Now())
		cv := colour.LightGreen.Sprintf(tag)

		exists := ver.TagExists(tag)
		if exists {
			fmt.Printf("Tag '%s' already exists\n", cv)
			return
		}

		fmt.Printf("Will create tag '%s' (hash %s)", cv, verifiedHash)
	},
}

var listTagCmd = &cobra.Command{
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

		if autoIncrement {
			nextInc, err := ver.GetLatestAutoInc(f)
			if err != nil {
				fmt.Printf("could not find next increment: %s\n", err.Error())
				os.Exit(1)
			}
			f.Modifier = f.Modifier + strconv.Itoa(nextInc)
		}

		tag := ""
		if len(args) > 0 {
			tag = args[0]
		}

		if tag == "" {
			tag, _ = f.Version(time.Now())
		}
		commit, err := ver.TagNext(ver.TagArgs{
			Hash: hash,
			Push: push,
			CV:   f,
			Tag:  tag,
		})
		CheckIfError(err)
		if lean {
			fmt.Println(tag)
		} else {
			fmt.Printf("Created tag '%s' (hash %s)\n", tag, commit)
		}
	},
}

var retagCmd = &cobra.Command{
	Use:   "retag",
	Short: "retag",
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

		tag := ""
		if len(args) > 0 {
			tag = args[0]
		}

		if tag == "" {
			tag, _ = f.Version(time.Now())
		}

		exists := ver.TagExists(tag)
		if !exists {
			CheckIfError(fmt.Errorf("tag '%s' does not exist", tag))
		}

		commit, err := ver.Retag(ver.TagArgs{
			Hash: hash,
			Push: push,
			CV:   f,
			Tag:  tag,
		})
		CheckIfError(err)
		fmt.Printf("Created tag '%s' (hash %s)\n", tag, commit)
	},
}

var untagCmd = &cobra.Command{
	Use:   "untag",
	Short: "untag",
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

		tag := ""
		if len(args) > 0 {
			tag = args[0]
		}

		if tag == "" {
			tag, _ = f.Version(time.Now())
		}

		exists := ver.TagExists(tag)
		if !exists {
			CheckIfError(fmt.Errorf("tag '%s' does not exist", tag))
		}

		err = ver.Untag(ver.TagArgs{
			Hash: hash,
			Push: push,
			CV:   f,
			Tag:  tag,
		})
		CheckIfError(err)
	},
}

func init() {
	rootCmd.AddCommand(listTagCmd)
	listTagCmd.Flags().BoolVar(&noColour, "no-colour", false, "Disable colour output")
	listTagCmd.Flags().BoolVar(&changelog, "changeLog", true, "Include changelog")
	listTagCmd.Flags().IntVarP(&limit, "limit", "l", 5, "Limit number of results (based on hashes)")

	rootCmd.AddCommand(latestTagCmd)
	latestTagCmd.Flags().BoolVar(&noColour, "no-colour", false, "Disable colour output")
	latestTagCmd.Flags().BoolVar(&changelog, "changeLog", true, "Include changelog")

	rootCmd.AddCommand(tagCmd)
	tagCmd.Flags().BoolVarP(&push, "push", "p", false, "Push tag after create")
	tagCmd.Flags().BoolVarP(&autoIncrement, "auto-increment", "i", false, "Adds an auto-incremented modifier, based off previous latest release")
	tagCmd.Flags().StringVar(&hash, "hash", "", "Override Hash")
	tagCmd.Flags().BoolVarP(&lean, "lean", "l", false, "Output the version number only")

	rootCmd.AddCommand(retagCmd)
	retagCmd.Flags().BoolVarP(&push, "push", "p", false, "Push tag after update")
	retagCmd.Flags().StringVar(&hash, "hash", "", "Override Hash")

	rootCmd.AddCommand(untagCmd)
	untagCmd.Flags().BoolVarP(&push, "push", "p", false, "Push tag after delete")
	untagCmd.Flags().StringVar(&hash, "hash", "", "Override Hash")

	rootCmd.AddCommand(nextTagCommand)
	nextTagCommand.Flags().StringVar(&hash, "hash", "HEAD", "Override Hash")
}
