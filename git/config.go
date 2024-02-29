package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/socialviolation/git-calver/ver"
)

func GetFormat() (*ver.Format, error) {
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

	f, err := ver.NewFormat(val)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func SetFormat(f *ver.Format) error {
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
