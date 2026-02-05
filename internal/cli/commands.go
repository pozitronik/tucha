package cli

// Command represents the CLI command to execute.
type Command int

// CLI commands.
const (
	CmdRun         Command = iota // Default: run server in foreground
	CmdHelp                       // Show help message
	CmdVersion                    // Show version and exit
	CmdBackground                 // Run server in background (daemon mode)
	CmdStatus                     // Show if server is running
	CmdStop                       // Stop background server
	CmdConfigCheck                // Validate configuration file
	CmdUserList                   // List users
	CmdUserAdd                    // Add user
	CmdUserRemove                 // Remove user
	CmdUserPwd                    // Set user password
	CmdUserQuota                  // Set user quota
	CmdUserInfo                   // Show user details
)

// Exit codes.
const (
	ExitSuccess        = 0 // Success
	ExitError          = 1 // General error
	ExitConfigError    = 2 // Configuration error
	ExitAlreadyRunning = 3 // Already running (--background)
	ExitNotRunning     = 4 // Not running (--stop/--status)
)

// HelpText returns the full help message.
func HelpText() string {
	return `Tucha - Cloud storage server

Usage: tucha [options] [command]

Options:
  -config <path>     Path to configuration file (default: config.yaml)

Commands:
  --help, --?        Show this help message
  --version          Show version and exit
  --background       Run server in background (daemon mode)
  --status           Show if server is running
  --stop             Stop background server
  --config-check     Validate configuration file

User Management:
  --user list [mask]               List users (optional email filter)
  --user add <email> <pwd> [quota] Add user (quota: "16GB", "512MB")
  --user remove <email>            Remove user
  --user pwd <email> <pwd>         Set password
  --user quota <email> <quota>     Set quota
  --user info <email>              Show user details

Examples:
  tucha                            Start in foreground
  tucha --background               Start in background
  tucha --user add user@x.com pass 8GB
  tucha --user list *@example.com
`
}
