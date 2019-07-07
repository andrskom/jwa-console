package tag

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/andrskom/jwa-console/pkg/config"
	"github.com/andrskom/jwa-console/pkg/timeline"
)

type Component struct {
	cfg *config.Component
}

func NewComponent(cfg *config.Component) *Component {
	return &Component{cfg: cfg}
}

func (c *Component) SetTag(tag string, noTag bool, m *timeline.Model) error {
	cfg, err := c.cfg.GetCfg()
	if err != nil {
		return err
	}
	if len(cfg.Tags) == 0 {
		return nil
	}
	if len(tag) > 0 && noTag {
		return errors.New("use either tag or no tag")
	}

	switch {
	case len(tag) > 0:
		hasTag := false
		for _, t := range cfg.Tags {
			if t == tag {
				hasTag = true
				break
			}
		}
		if !hasTag {
			return errors.New("u set unexpected tag")
		}
		m.Tag = tag
	case noTag:
		return nil
	default:
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Please, choose tag:\n")
		fmt.Print("[nt] no tag\n")
		for i, t := range cfg.Tags {
			fmt.Printf("[%d] %s\n", i, t)
		}
		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		if text == "nt" {
			return nil
		}
		i, err := strconv.Atoi(text)
		if err != nil {
			return err
		}
		if i < 0 || i >= len(cfg.Tags) {
			return errors.New("wrong index of tag")
		}
		m.Tag = cfg.Tags[i]
	}

	return nil
}
