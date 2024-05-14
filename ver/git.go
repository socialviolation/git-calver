package ver

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func GetRepoFormat() (*Format, bool, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, false, fmt.Errorf("could not init repo at .: %w", err)
	}

	conf, err := r.Config()
	if err != nil {
		return nil, false, fmt.Errorf("could not retrieve config: %w", err)
	}

	if !conf.Raw.HasSection("calver") {
		return nil, false, fmt.Errorf("[calver] not set")
	}

	val := conf.Raw.Section("calver").Option("format")
	if val == "" {
		return nil, false, fmt.Errorf("[calver].format not set")
	}

	f, err := NewFormat(val)
	if err != nil {
		return nil, false, err
	}
	return f, strings.HasSuffix(val, "-A"), nil
}

func SetRepoFormat(f *Format) error {
	r, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("could not init repo at .: %w", err)
	}

	conf, err := r.Config()
	if err != nil {
		return fmt.Errorf("could not retrieve config: %w", err)
	}
	if err != nil {
		return err
	}

	conf.Raw.SetOption("calver", "", "format", f.String())
	err = r.SetConfig(conf)
	if err != nil {
		return err
	}

	return nil
}

type TagArgs struct {
	CV   *CalVer
	Hash string
	Push bool
	Tag  string
}

func LatestTag(reg *regexp.Regexp, changelog bool) (*CalVerTagGroup, error) {
	latestList, err := ListTags(reg, 1, changelog)
	if err != nil {
		return nil, err
	}
	for _, tagGroup := range latestList {
		return tagGroup, nil
	}

	return nil, fmt.Errorf("no latest tag found")
}

func GetLatestAutoInc(cv *CalVer) (int, error) {
	latest, err := LatestTag(cv.Regex(), false)
	autoMod := 1
	if err != nil {
		return 0, fmt.Errorf("could not find latest tag with format: %s", cv.Format.String())
	}

	nextTagStr, err := cv.Version(time.Now())
	if err != nil {
		return 0, fmt.Errorf("could not generate next version: %w", err)
	}

	nextCalVer := strings.Split(nextTagStr, "-")[0]
	foundIncableV := false

	for _, tag := range latest.Tags {
		if !strings.HasPrefix(tag, nextCalVer) {
			// No matching tag found to increment
			continue
		}

		foundIncableV = true
		latestModBits := strings.Split(tag, "-")
		if len(latestModBits) <= 1 {
			autoMod = 1
			break
		}

		latestMod := latestModBits[1]
		if strings.HasPrefix(latestMod, cv.Modifier) {
			incPartMod := strings.Replace(latestMod, cv.Modifier, "", 1)
			oldInc, err := strconv.Atoi(incPartMod)
			if err != nil {
				continue
			}
			autoMod = oldInc + 1
		}
	}

	if !foundIncableV {
		return 1, nil
	}

	return autoMod, nil
}

func ListTags(reg *regexp.Regexp, limit int, changelog bool) ([]*CalVerTagGroup, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("could not init repo at .: %w", err)
	}

	refs, err := r.Tags()
	if err != nil {
		return nil, fmt.Errorf("could not find ags: %w", err)
	}

	tagMap := make(map[string]*CalVerTagGroup)
	hashes := make([]string, 0)
	err = refs.ForEach(func(tag *plumbing.Reference) error {
		short := tag.Name().Short()
		if !reg.Match([]byte(short)) {
			return nil
		}
		co, _ := getCommitByTag(r, string(tag.Name()))
		if co == nil {
			return nil
		}

		hash := co.Hash.String()[:7]
		inc := func(l []string, s string) bool {
			for _, c := range l {
				if c == s {
					return true
				}
			}
			return false
		}
		if !inc(hashes, hash) {
			hashes = append(hashes, hash)
		}

		if tagMap[hash] == nil {
			tagMap[hash] = &CalVerTagGroup{
				Hash:      hash,
				Commit:    co,
				When:      co.Author.When,
				Tags:      []string{short},
				Refs:      []*plumbing.Reference{tag},
				ChangeLog: []*object.Commit{co},
			}
			return nil
		}

		tagMap[hash].Tags = append(tagMap[hash].Tags, short)
		tagMap[hash].Refs = append(tagMap[hash].Refs, tag)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(hashes, func(i, j int) bool {
		return tagMap[hashes[j]].Time().Before(tagMap[hashes[i]].Time())
	})

	if changelog {
		changeLimit := 20
		for i, hash := range hashes {
			since := time.Time{}
			if i < len(hashes)-2 {
				since = tagMap[hashes[i+1]].Commit.Author.When.Add(time.Second * 1)
			}

			logs, _ := r.Log(&git.LogOptions{
				Order: git.LogOrderCommitterTime,
				Since: &since,
				Until: &tagMap[hash].Commit.Author.When,
			})
			if logs == nil {
				continue
			}
			includes := func(l []*object.Commit, commit *object.Commit) bool {
				for _, c := range l {
					if c.Hash.String() == commit.Hash.String() {
						return true
					}
				}
				return false
			}
			_ = logs.ForEach(func(commit *object.Commit) error {
				if !includes(tagMap[hash].ChangeLog, commit) {
					tagMap[hash].ChangeLog = append(tagMap[hash].ChangeLog, commit)
				}
				return nil
			})
			changeLimit++
			if changeLimit > 20 {
				break
			}
		}
	}

	results := make([]*CalVerTagGroup, 0)
	for i, hash := range hashes {
		if i == 0 {
			tagMap[hash].Latest = true
		}
		if i >= limit {
			break
		}
		results = append(results, tagMap[hash])
	}

	return results, nil
}

func VerifyHash(hash string) (string, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("could not init repo at .: %w", err)
	}
	if hash == "" || hash == "HEAD" {
		head, err := r.Head()
		if err != nil {
			return "", fmt.Errorf("cannot get HEAD: %w", err)
		}
		return head.Hash().String(), nil
	} else if len(hash) < 8 {
		co, _ := findShortHash(r, hash)
		hash = co.Hash.String()
	}

	co, err := r.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return "", fmt.Errorf("cannot find hash %s", hash)
	}
	return co.Hash.String(), nil
}

func TagNext(args TagArgs) (string, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("could not init repo at .: %w", err)
	}

	v := args.Tag
	if v == "" {
		v, err = args.CV.Version(time.Now())
		if err != nil {
			return "", err
		}
	}

	var co *object.Commit
	if args.Hash == "" {
		h, err := r.ResolveRevision("HEAD")
		if err != nil {
			return "", fmt.Errorf("could not resolve HEAD: %w", err)
		}
		co, err = r.CommitObject(*h)
		if err != nil {
			return "", err
		}
	} else {
		co, err = findShortHash(r, args.Hash)
		if err != nil {
			return "", err
		}
	}

	_, err = setTag(r, v, co)
	if err != nil {
		return "", fmt.Errorf("could not create tag: %w", err)
	}

	if args.Push {
		err = pushTags(false, v)
	}
	if co == nil {
		return "HEAD", err
	}

	return co.Hash.String()[:7], nil
}

func Untag(args TagArgs) error {
	r, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("could not init repo at .: %w", err)
	}

	oldHash := tagHash(args.Tag)
	err = r.DeleteTag(args.Tag)
	if err != nil {
		return fmt.Errorf("could not delete tag '%s': %w", args.Tag, err)
	}

	fmt.Printf("Deleted tag '%s' (hash %s)\n", args.Tag, oldHash[:7])

	if args.Push {
		return pushTags(true, args.Tag)
	}
	return nil
}

func Retag(args TagArgs) (string, error) {
	err := Untag(args)
	if err != nil {
		return "", err
	}

	return TagNext(args)
}

func TagExists(tag string) bool {
	r, err := git.PlainOpen(".")
	if err != nil {
		return false
	}
	return tagExists(r, tag)
}

func findShortHash(r *git.Repository, hash string) (*object.Commit, error) {
	cos, err := r.CommitObjects()
	if err != nil {
		return nil, err
	}

	var c *object.Commit
	err = cos.ForEach(func(commit *object.Commit) error {
		if strings.HasPrefix(commit.Hash.String(), hash) {
			c = commit
			return fmt.Errorf("found")
		}
		return nil
	})
	if err != nil && err.Error() != "found" {
		return nil, err
	}

	return c, nil
}

func getCommitByTag(r *git.Repository, tag string) (*object.Commit, error) {
	rev, _ := r.ResolveRevision(plumbing.Revision(tag))
	co, err := r.CommitObject(plumbing.NewHash(rev.String()))
	if err != nil {
		return nil, err
	}
	return co, nil
}

func tagHash(tag string) string {
	r, err := git.PlainOpen(".")
	if err != nil {
		return ""
	}

	ref, err := r.Tag(tag)
	if err != nil {
		return ""
	}
	co, err := getCommitByTag(r, string(ref.Name()))
	if err != nil {
		return ""
	}
	return co.Hash.String()
}

func tagExists(r *git.Repository, tag string) bool {
	tagFoundErr := "tag was found"
	tags, err := r.Tags()
	if err != nil {
		log.Printf("get tags error: %s", err)
		return false
	}
	res := false
	err = tags.ForEach(func(t *plumbing.Reference) error {
		if t.Name().Short() == tag {
			res = true
			return fmt.Errorf(tagFoundErr)
		}
		return nil
	})
	if err != nil && err.Error() != tagFoundErr {
		fmt.Printf("iterate ags error: %s", err)
		return false
	}
	return res
}

func setTag(r *git.Repository, tag string, co *object.Commit) (bool, error) {
	if tagExists(r, tag) {
		fmt.Printf("tag %s already exists\n", tag)
		return false, nil
	}
	if co == nil {
		h, err := r.Head()
		if err != nil {
			fmt.Printf("get HEAD error: %s", err)
			return false, err
		}
		co, err = findShortHash(r, h.Hash().String())
		if err != nil {
			fmt.Printf("get hash (%s) error: %s", h.Hash(), err)
			return false, err
		}
	}

	gitPathCmd := exec.Command("which", "git")
	gitPath, err := gitPathCmd.Output()
	if err != nil {
		return false, fmt.Errorf("could not find git on host system, push is not supported")
	}

	gitCmd := strings.Replace(string(gitPath), "\n", "", -1)
	cmd := exec.Command(gitCmd, "tag", tag)
	_, err = cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
	}

	if err != nil {
		fmt.Printf("create tag error: %s\n", err)
		return false, err
	}

	return true, nil
}

func pushTags(deletion bool, tag ...string) error {
	if len(tag) == 0 {
		return fmt.Errorf("no tags to push")
	}
	gitPathCmd := exec.Command("which", "git")
	gitPath, err := gitPathCmd.Output()
	if err != nil {
		return fmt.Errorf("could not find git on host system, push is not supported")
	}

	gitCmd := strings.Replace(string(gitPath), "\n", "", -1)

	for _, t := range tag {
		action := fmt.Sprintf("pushing %s to origin\n", t)
		cmd := exec.Command(gitCmd, "push", "origin", t)

		if deletion {
			cmd = exec.Command(gitCmd, "push", "origin", "--delete", t)
			action = fmt.Sprintf("deleting %s on origin\n", t)
		}
		fmt.Printf(action)
		stdout, err := cmd.Output()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		fmt.Println(string(stdout))
	}

	return nil
}
