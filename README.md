# jwa-console

Jira worklog assistant as console application.
Also tray util.

```
WIP!!! It's a Alpha version of app! Not ready for production!
```

## Supported OS.

   - mac os
   - linux

## Instruction.

### Installation.

Download from github releases and put it to $PATH.
Or use make file if u have installed golang with cmd 
`make install`(your $GOBIN must be in $PATH).

### Getting started.

1. First time u must init application. `jwac init`
2. Second time u must login to jira. `jwac login https://your_jira_domain/`
3. Add to you console rc file(for example `$HOME/.zshrc` if u like zsh)
row `source <(jwac completion zsh)`. As result you will have completion for jwac.
4. Use `jwac help` for learning application.

## Tray util.

Simple run tray util from arch or build with go 
and u can see active or not task in your tray.

Example:
```
jwac-tray &
```

You can stop it with kill util)))

## Help

If u want to help me with your PR or Issue, i will be very happy.
