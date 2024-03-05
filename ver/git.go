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
	"strings"
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

func ListTags(format *Format, limit int) (map[string][]CalVerTag, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("could not init repo at .: %w", err)
	}

	refs, err := r.Tags()
	if err != nil {
		return nil, fmt.Errorf("could not find tags: %w", err)
	}
	tags := make([]CalVerTag, 0)
	regex := format.Regex()

	err = refs.ForEach(func(tag *plumbing.Reference) error {
		if !tag.Name().IsTag() {
			return nil
		}

		short := tag.Name().Short()
		if !regex.Match([]byte(short)) {
			return nil
		}
		rev, _ := r.ResolveRevision(plumbing.Revision(tag.Name()))
		co, _ := r.CommitObject(plumbing.NewHash(rev.String()))
		if co == nil {
			return nil
		}

		tags = append(tags, CalVerTag{
			Tag:    tag,
			Commit: co,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	//sort.Slice(tags, func(i, j int) bool {
	//	return tags[j].Time().Before(tags[i].Time())
	//})

	inc := 0
	tagMap := make(map[string][]CalVerTag)
	for _, tag := range tags {
		hash := tag.Tag.Hash().String()[:7]
		if tagMap[hash] == nil {
			tagMap[hash] = []CalVerTag{tag}
			inc++
			if inc >= limit {
				break
			}
			continue
		}

		tagMap[hash] = append(tagMap[hash], tag)
	}
	return tagMap, nil
}

func TagNext(ver CalVer, hash string) error {
	r, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("could not init repo at .: %w", err)
	}

	v, err := ver.Version(time.Now())
	if err != nil {
		return err
	}

	if hash == "" {
		h, err := r.ResolveRevision("HEAD")
		if err != nil {
			return fmt.Errorf("could not resolve HEAD: %w", err)
		}
		hash = h.String()
	} else {
		_, err := r.ResolveRevision(plumbing.Revision(hash))
		if err != nil {
			return fmt.Errorf("could not resolve hash %s: %w", hash, err)
		}
	}
	_, err = setTag(r, v)
	if err != nil {
		return fmt.Errorf("could not create tag: %w", err)
	}
	return nil
}

func Retag(ver CalVer, hash string) error {
	r, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("could not init repo at .: %w", err)
	}

	v, err := ver.Version(time.Now())
	if err != nil {
		return err
	}

	if hash == "" {
		hash = "HEAD"
	}
	_, err = r.CreateTag(v, plumbing.NewHash(hash), &git.CreateTagOptions{
		//Force: true,
	})
	if err != nil {
		return fmt.Errorf("could not create tag: %w", err)
	}
	return nil
}

func tagExists(r *git.Repository, tag string) bool {
	tagFoundErr := "tag was found"
	tags, err := r.TagObjects()
	if err != nil {
		log.Printf("get tags error: %s", err)
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
		log.Printf("iterate tags error: %s", err)
		return false
	}
	return res
}

func setTag(r *git.Repository, tag string) (bool, error) {
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
		Message: tag,
	})

	if err != nil {
		log.Printf("create tag error: %s", err)
		return false, err
	}

	return true, nil
}

func pushTags(r *git.Repository) error {
	po := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
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

type CalVerTag struct {
	Tag    *plumbing.Reference
	Commit *object.Commit
}

func (cvt CalVerTag) Time() time.Time {
	if cvt.Commit == nil {
		return time.Time{}
	}
	return cvt.Commit.Author.When
}

func (cvt CalVerTag) Surname() string {
	if cvt.Commit == nil {
		return ""
	}

	b := strings.Split(cvt.Commit.Author.Name, " ")
	return b[len(b)-1]
}
