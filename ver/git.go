package ver

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"os"
	"sort"
	"time"
)

func GetRepoFormat() (*Format, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("could not init repo at .: %w", err)
	}

	conf, err := r.Config()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve config: %w", err)
	}
	if err != nil {
		return nil, err
	}

	if !conf.Raw.HasSection("calver") {
		return nil, fmt.Errorf("[calver] not set")
	}

	val := conf.Raw.Section("calver").Option("format")
	if val == "" {
		return nil, fmt.Errorf("[calver].format not set")
	}

	f, err := NewFormat(val)
	if err != nil {
		return nil, err
	}
	return f, nil
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

func LatestTag(format *Format, changelog bool) (*CalVerTagGroup, error) {
	latestList, err := ListTags(format, 1, changelog)
	if err != nil {
		return nil, err
	}
	for _, tagGroup := range latestList {
		return tagGroup, nil
	}

	return nil, fmt.Errorf("no latest tag found")
}

func ListTags(format *Format, limit int, changelog bool) ([]*CalVerTagGroup, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("could not init repo at .: %w", err)
	}

	refs, err := r.Tags()
	if err != nil {
		return nil, fmt.Errorf("could not find printTags: %w", err)
	}
	regex := format.Regex()

	tagMap := make(map[string]*CalVerTagGroup)
	hashes := make([]string, 0)
	err = refs.ForEach(func(tag *plumbing.Reference) error {
		short := tag.Name().Short()
		if !regex.Match([]byte(short)) {
			return nil
		}
		rev, _ := r.ResolveRevision(plumbing.Revision(tag.Name()))
		co, _ := r.CommitObject(plumbing.NewHash(rev.String()))
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

type TagArgs struct {
	CV   *CalVer
	Hash string
	Push bool
	Tag  string
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
		return head.Hash().String()[:7], nil
	}
	co, err := r.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return "", fmt.Errorf("cannot find hash %s", hash)
	}
	return co.Hash.String()[:7], nil
}

func TagNext(args TagArgs) (string, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("could not init repo at .: %w", err)
	}

	v, err := args.CV.Version(time.Now())
	if err != nil {
		return "", err
	}

	hRaw := plumbing.NewHash(args.Hash)
	h := &hRaw
	if h.IsZero() {
		h, err = r.ResolveRevision("HEAD")
		if err != nil {
			return "", fmt.Errorf("could not resolve HEAD: %w", err)
		}
	} else {
		h, err = r.ResolveRevision(plumbing.Revision(h.String()))
		if err != nil {
			return "", fmt.Errorf("could not resolve hash %s: %w", h.String(), err)
		}
	}

	co, err := r.CommitObject(*h)
	created, err := setTag(r, v, co)
	if err != nil {
		return "", fmt.Errorf("could not create tag: %w", err)
	}

	if created && args.Push {
		err = pushTags(r, v)
	}
	return co.Hash.String()[:7], nil
}

//func Retag(ver CalVer, hash string) error {
//	r, err := git.PlainOpen(".")
//	if err != nil {
//		return fmt.Errorf("could not init repo at .: %w", err)
//	}
//
//	v, err := ver.Version(time.Now())
//	if err != nil {
//		return err
//	}
//
//	if hash == "" {
//		hash = "HEAD"
//	}
//	_, err = r.CreateTag(v, plumbing.NewHash(hash), &git.CreateTagOptions{
//		//Force: true,
//	})
//	if err != nil {
//		return fmt.Errorf("could not create tag: %w", err)
//	}
//	return nil
//}

func TagExists(tag string) bool {
	r, err := git.PlainOpen(".")
	if err != nil {
		return false
	}
	return tagExists(r, tag)
}

func tagExists(r *git.Repository, tag string) bool {
	tagFoundErr := "tag was found"
	tags, err := r.TagObjects()
	if err != nil {
		log.Printf("get printTags error: %s", err)
		return false
	}
	res := false
	err = tags.ForEach(func(t *object.Tag) error {
		if t.Name == tag {
			res = true
			return fmt.Errorf(tagFoundErr)
		}
		return nil
	})
	if err != nil && err.Error() != tagFoundErr {
		log.Printf("iterate printTags error: %s", err)
		return false
	}
	return res
}

func setTag(r *git.Repository, tag string, co *object.Commit) (bool, error) {
	if tagExists(r, tag) {
		log.Printf("tag %s already exists", tag)
		return false, nil
	}
	h, err := r.Head()
	if err != nil {
		log.Printf("get HEAD error: %s", err)
		return false, err
	}

	_, err = r.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Tagger:  &co.Author,
		Message: tag,
	})

	if err != nil {
		log.Printf("create tag error: %s", err)
		return false, err
	}

	return true, nil
}

func pushTags(r *git.Repository, tag ...string) error {
	if len(tag) == 0 {
		return fmt.Errorf("no printTags to push")
	}
	refSpecs := make([]config.RefSpec, len(tag))
	for _, t := range tag {
		refSpecs = append(refSpecs, config.RefSpec(fmt.Sprintf("refs/printTags/%s:refs/printTags/%s", t, t)))
	}

	po := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs:   refSpecs,
	}
	err := r.Push(po)

	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			log.Print("origin remote was up to date, no push done")
			return nil
		}
		log.Printf("push to remote origin error: %s", err)
		return err
	}

	return nil
}
