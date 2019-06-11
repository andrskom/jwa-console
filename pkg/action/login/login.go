package login

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/andrskom/jwa-console/pkg/creds"
)

func Login(credsComponent *creds.Component) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		addr := c.Args().First()
		if addr == "" {
			return errors.New("u must set host of jira as last arg")
		}
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Username:")
		login, err := reader.ReadBytes('\n')
		if err != nil {
			return err
		}
		fmt.Println("Password:")
		pass, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			return err
		}

		model := creds.Model{
			Username: strings.TrimRight(string(login), "\n"),
			Addr:     addr,
			Password: string(pass),
		}

		tp := jira.BasicAuthTransport{
			Username: model.Username,
			Password: model.Password,
		}

		jiraClient, err := jira.NewClient(tp.Client(), model.Addr)
		if err != nil {
			return err
		}

		if _, resp, err := jiraClient.User.GetSelf(); err != nil {
			 return fmt.Errorf("unexpected error code from jira: %d", resp.StatusCode)
		}

		return credsComponent.Save(&model)
	}
}
