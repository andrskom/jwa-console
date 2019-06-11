package action

import (
	"errors"
	"fmt"

	"github.com/urfave/cli"
)

func Completion() func(c *cli.Context) error {
	bash := `
: ${PROG:=jwac}

_cli_bash_autocomplete() {
    local cur opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}

complete -F _cli_bash_autocomplete $PROG

unset PROG`

	zsh := `
: ${PROG:=jwac}

_cli_bash_autocomplete() {
    local cur opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}

complete -F _cli_bash_autocomplete $PROG

unset PROG`

	return func(c *cli.Context) error {
		switch c.Args().Get(0) {
		case "zsh":
			fmt.Println(zsh)
		case "bash":
			fmt.Println(bash)
		default:
			return errors.New("unexpected type of completion")
		}
		return nil
	}
}
