package regloader

import (
	"path/filepath"

	"github.com/fatih/structs"
	"github.com/leeola/errors"
	"github.com/leeola/fixity"
	"github.com/leeola/fixity/autoload/registry"
	"github.com/leeola/fixity/fixities/local"
	"github.com/leeola/fixity/indexes/snail"
	cu "github.com/leeola/fixity/util/configunmarshaller"
	homedir "github.com/mitchellh/go-homedir"
)

func init() {
	registry.RegisterIndex(Loader)
}

func Loader(cu cu.ConfigUnmarshaller) (fixity.Index, error) {
	var c struct {
		DontExpandHome bool         `toml:"dontExpandHome"`
		LocalConfig    local.Config `toml:"localFixity"`
		Config         snail.Config `toml:"snailIndex"`
	}

	if err := cu.Unmarshal(&c); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	// if the config isn't defined, do not load anything. This is allowed.
	if structs.IsZero(c.Config) {
		return nil, nil
	}

	if !c.DontExpandHome {
		if c.LocalConfig.RootPath != "" {
			p, err := homedir.Expand(c.LocalConfig.RootPath)
			if err != nil {
				return nil, err
			}
			c.LocalConfig.RootPath = p
		}

		if c.Config.Path != "" {
			p, err := homedir.Expand(c.Config.Path)
			if err != nil {
				return nil, err
			}
			c.Config.Path = p
		}
	}

	// don't join with the root path if the path is empty or absolute.
	if c.Config.Path != "" && !filepath.IsAbs(c.Config.Path) {
		// rootPath being empty is okay, Join ignores empty.
		c.Config.Path = filepath.Join(c.LocalConfig.RootPath, c.Config.Path)
	}

	return snail.New(c.Config)
}
