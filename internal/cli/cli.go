package cli

import (
	"fmt"
	"strings"
)

// CLI holds the parsed command-line arguments.
type CLI struct {
	ConfigPath string   // Path to configuration file
	Command    Command  // The command to execute
	Args       []string // Additional arguments for the command
}

// Parse parses command-line arguments and returns a CLI struct.
// args should be os.Args (including program name at index 0).
func Parse(args []string) (*CLI, error) {
	cli := &CLI{
		ConfigPath: "config.yaml",
		Command:    CmdRun,
		Args:       []string{},
	}

	if len(args) < 1 {
		return cli, nil
	}

	// Skip program name
	args = args[1:]

	i := 0
	for i < len(args) {
		arg := args[i]

		switch {
		case arg == "-config" || arg == "--config":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			i++
			cli.ConfigPath = args[i]

		case arg == "--help" || arg == "--?" || arg == "-help" || arg == "-?" || arg == "-h":
			cli.Command = CmdHelp
			return cli, nil

		case arg == "--version" || arg == "-version" || arg == "-v":
			cli.Command = CmdVersion
			return cli, nil

		case arg == "--background" || arg == "-background":
			cli.Command = CmdBackground
			return cli, nil

		case arg == "--status" || arg == "-status":
			cli.Command = CmdStatus
			return cli, nil

		case arg == "--stop" || arg == "-stop":
			cli.Command = CmdStop
			return cli, nil

		case arg == "--config-check" || arg == "-config-check":
			cli.Command = CmdConfigCheck
			return cli, nil

		case arg == "--user" || arg == "-user":
			return parseUserCommand(cli, args[i+1:])

		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown option: %s", arg)
			}
			// Non-option argument, stop parsing
			cli.Args = args[i:]
			return cli, nil
		}

		i++
	}

	return cli, nil
}

// parseUserCommand parses the --user subcommand.
func parseUserCommand(cli *CLI, args []string) (*CLI, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("--user requires a subcommand (list, add, remove, pwd, quota, sizelimit, history, info)")
	}

	subCmd := strings.ToLower(args[0])
	rest := args[1:]

	switch subCmd {
	case "list":
		cli.Command = CmdUserList
		cli.Args = rest // Optional mask

	case "add":
		cli.Command = CmdUserAdd
		if len(rest) < 2 {
			return nil, fmt.Errorf("--user add requires <email> <password> [quota]")
		}
		cli.Args = rest // email, password, [quota]

	case "remove":
		cli.Command = CmdUserRemove
		if len(rest) < 1 {
			return nil, fmt.Errorf("--user remove requires <email>")
		}
		cli.Args = rest // email

	case "pwd":
		cli.Command = CmdUserPwd
		if len(rest) < 2 {
			return nil, fmt.Errorf("--user pwd requires <email> <password>")
		}
		cli.Args = rest // email, password

	case "quota":
		cli.Command = CmdUserQuota
		if len(rest) < 2 {
			return nil, fmt.Errorf("--user quota requires <email> <quota>")
		}
		cli.Args = rest // email, quota

	case "sizelimit":
		cli.Command = CmdUserSizeLimit
		if len(rest) < 2 {
			return nil, fmt.Errorf("--user sizelimit requires <email> <size>")
		}
		cli.Args = rest // email, size

	case "history":
		cli.Command = CmdUserHistory
		if len(rest) < 2 {
			return nil, fmt.Errorf("--user history requires <email> <on|off>")
		}
		cli.Args = rest // email, on|off

	case "info":
		cli.Command = CmdUserInfo
		if len(rest) < 1 {
			return nil, fmt.Errorf("--user info requires <email>")
		}
		cli.Args = rest // email

	default:
		return nil, fmt.Errorf("unknown --user subcommand: %s", subCmd)
	}

	return cli, nil
}
