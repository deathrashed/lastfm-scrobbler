package completion

import (
	"fmt"
	"strings"
)

type Shell string

const (
	ShellZsh        Shell = "zsh"
	ShellBash       Shell = "bash"
	ShellFish       Shell = "fish"
	ShellPowerShell Shell = "powershell"
)

var Shells = []Shell{ShellZsh, ShellBash, ShellFish, ShellPowerShell}

func ParseShell(value string) (Shell, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ShellZsh):
		return ShellZsh, nil
	case string(ShellBash):
		return ShellBash, nil
	case string(ShellFish):
		return ShellFish, nil
	case string(ShellPowerShell), "pwsh":
		return ShellPowerShell, nil
	default:
		return "", fmt.Errorf("unsupported shell %q; choose zsh, bash, fish, or powershell", shellDisplay(value))
	}
}

func (s Shell) String() string { return string(s) }

type completionFlag struct {
	Name        string
	Description string
	Value       bool
}

type completionCommand struct {
	Name        string
	Description string
	Flags       []completionFlag
}

var completionCommands = []completionCommand{
	{Name: "tui", Description: "launch the terminal interface"},
	{Name: "setup", Description: "run the first-time setup wizard"},
	{Name: "manual", Description: "scrobble one Artist - Album", Flags: []completionFlag{{Name: "loop", Description: "album loops", Value: true}, {Name: "limit", Description: "tracks per album", Value: true}, {Name: "interval", Description: "delay", Value: true}, {Name: "dry-run", Description: "do not scrobble"}, {Name: "json", Description: "JSON output"}, {Name: "artist", Description: "artist name", Value: true}, {Name: "album", Description: "album name", Value: true}}},
	{Name: "file", Description: "import a list, playlist, or folder", Flags: []completionFlag{{Name: "loop", Description: "album loops", Value: true}, {Name: "limit", Description: "tracks per album", Value: true}, {Name: "interval", Description: "delay", Value: true}, {Name: "dry-run", Description: "do not scrobble"}, {Name: "json", Description: "JSON output"}}},
	{Name: "discography", Description: "list or scrobble Last.fm top albums", Flags: []completionFlag{{Name: "loop", Description: "album loops", Value: true}, {Name: "limit", Description: "tracks per album", Value: true}, {Name: "interval", Description: "delay", Value: true}, {Name: "dry-run", Description: "do not scrobble"}, {Name: "json", Description: "JSON output"}, {Name: "all", Description: "select all albums"}, {Name: "albums", Description: "album names", Value: true}, {Name: "first", Description: "first albums", Value: true}, {Name: "clean", Description: "remove noisy releases"}}},
	{Name: "similar", Description: "list similar album suggestions", Flags: []completionFlag{{Name: "limit", Description: "result count", Value: true}, {Name: "json", Description: "JSON output"}}},
	{Name: "test", Description: "test API and authentication", Flags: []completionFlag{{Name: "json", Description: "JSON output"}}},
	{Name: "diagnostics", Description: "export a redacted support bundle", Flags: []completionFlag{{Name: "json", Description: "JSON output"}}},
	{Name: "check-update", Description: "check the configured update source", Flags: []completionFlag{{Name: "json", Description: "JSON output"}}},
	{Name: "completion", Description: "print or install shell completion"},
	{Name: "version", Description: "print version information"},
	{Name: "help", Description: "print command help"},
}

func Generate(shell string) (string, error) {
	parsed, err := ParseShell(shell)
	if err != nil {
		return "", err
	}
	switch parsed {
	case ShellZsh:
		return renderZsh(), nil
	case ShellBash:
		return renderBash(), nil
	case ShellFish:
		return renderFish(), nil
	case ShellPowerShell:
		return renderPowerShell(), nil
	default:
		return "", fmt.Errorf("unsupported shell %q", shell)
	}
}

func renderZsh() string {
	var b strings.Builder
	b.WriteString("#compdef scrobbler\n_scrobbler() {\n  local -a commands\n  commands=(\n")
	for _, command := range completionCommands {
		fmt.Fprintf(&b, "    '%s:%s'\n", command.Name, command.Description)
	}
	b.WriteString("  )\n  _arguments -C \\\n    '1:command:->command' \\\n    '*::argument:->args'\n  case $state in\n    command) _describe 'command' commands ;;\n    args)\n      case $words[2] in\n")
	for _, command := range completionCommands {
		if command.Name == "completion" {
			b.WriteString("        completion) _values 'shell' install zsh bash fish powershell pwsh ;;\n")
			continue
		}
		if len(command.Flags) == 0 {
			continue
		}
		fmt.Fprintf(&b, "        %s) _arguments", command.Name)
		for _, flag := range command.Flags {
			if flag.Value {
				fmt.Fprintf(&b, " '--%s[%s]:value'", flag.Name, flag.Description)
			} else {
				fmt.Fprintf(&b, " '--%s[%s]'", flag.Name, flag.Description)
			}
		}
		b.WriteString(" ;;\n")
	}
	b.WriteString("      esac\n    ;;\n  esac\n}\n_scrobbler \"$@\"\n")
	return b.String()
}

func renderBash() string {
	var b strings.Builder
	b.WriteString("_scrobbler_complete() {\n  local cur commands\n  COMPREPLY=()\n  cur=\"${COMP_WORDS[COMP_CWORD]}\"\n  commands=\"")
	for index, command := range completionCommands {
		if index > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(command.Name)
	}
	b.WriteString("\"\n  if [[ ${COMP_CWORD} -eq 1 ]]; then\n    COMPREPLY=( $(compgen -W \"${commands}\" -- \"${cur}\") )\n    return\n  fi\n  case \"${COMP_WORDS[1]}\" in\n")
	for _, command := range completionCommands {
		if command.Name == "completion" {
			b.WriteString("    completion) COMPREPLY=( $(compgen -W \"install zsh bash fish powershell pwsh\" -- \"${cur}\") ) ;;\n")
			continue
		}
		if len(command.Flags) == 0 {
			continue
		}
		fmt.Fprintf(&b, "    %s) COMPREPLY=( $(compgen -W \"", command.Name)
		for index, flag := range command.Flags {
			if index > 0 {
				b.WriteByte(' ')
			}
			fmt.Fprintf(&b, "--%s", flag.Name)
		}
		b.WriteString("\" -- \"${cur}\") ) ;;\n")
	}
	b.WriteString("  esac\n}\ncomplete -F _scrobbler_complete scrobbler\n")
	return b.String()
}

func renderFish() string {
	var b strings.Builder
	b.WriteString("complete -c scrobbler -f\n")
	for _, command := range completionCommands {
		fmt.Fprintf(&b, "complete -c scrobbler -n '__fish_use_subcommand' -a %s -d '%s'\n", command.Name, command.Description)
	}
	for _, command := range completionCommands {
		if command.Name == "completion" {
			b.WriteString("complete -c scrobbler -n '__fish_seen_subcommand_from completion' -a 'install zsh bash fish powershell pwsh'\n")
		}
		for _, flag := range command.Flags {
			fmt.Fprintf(&b, "complete -c scrobbler -n '__fish_seen_subcommand_from %s' -l %s", command.Name, flag.Name)
			if flag.Value {
				b.WriteString(" -r")
			}
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func renderPowerShell() string {
	return `$scrobblerCommands = @('tui', 'setup', 'manual', 'file', 'discography', 'similar', 'test', 'diagnostics', 'check-update', 'completion', 'version', 'help')
$scrobblerFlags = @('--loop', '--limit', '--interval', '--dry-run', '--json', '--artist', '--album', '--all', '--albums', '--first', '--clean')
Register-ArgumentCompleter -Native -CommandName scrobbler -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $elements = $commandAst.CommandElements
    if ($elements.Count -le 1) {
        $scrobblerCommands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
        }
        return
    }
    if ($elements[1].Value -eq 'completion' -and $elements.Count -le 3) {
        'install', 'zsh', 'bash', 'fish', 'powershell', 'pwsh' | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
        return
    }
    $scrobblerFlags | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
    }
}
`
}

func shellDisplay(value string) string {
	return strings.TrimSpace(value)
}
