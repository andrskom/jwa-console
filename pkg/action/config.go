package action

import (
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/andrskom/jwa-console/pkg/config"
)

func Config(
	cfg *config.Component,
) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.Bool("l") {
			model, err := cfg.GetCfg()
			if err != nil {
				return err
			}
			for k, v := range model.AsMap() {
				fmt.Printf("%20s | '%s'\n", k, v)
			}
			return nil
		}
		if len(c.String("set")) > 0 {
			data := c.String("set")
			kv := strings.Split(data, ":")
			if len(kv) != 2 {
				return errors.New("you must use ':' as separator for key and value")
			}

			model, err := cfg.GetCfg()
			if err != nil {
				return err
			}
			if err := model.Set(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])); err != nil {
				return err
			}
			return cfg.Save(model)
		}
		return nil
	}
}
