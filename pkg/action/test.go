package action

import (
	"log"

	"github.com/urfave/cli"
)

func Test() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		log.Printf("%#v", )
		return nil
	}
}
