package cli

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCmd    Command
		wantConfig string
		wantArgs   []string
		wantErr    bool
	}{
		// Default behavior
		{
			name:       "no args",
			args:       []string{"tucha"},
			wantCmd:    CmdRun,
			wantConfig: "config.yaml",
			wantArgs:   []string{},
		},

		// Help variants
		{
			name:    "help with --help",
			args:    []string{"tucha", "--help"},
			wantCmd: CmdHelp,
		},
		{
			name:    "help with --?",
			args:    []string{"tucha", "--?"},
			wantCmd: CmdHelp,
		},
		{
			name:    "help with -h",
			args:    []string{"tucha", "-h"},
			wantCmd: CmdHelp,
		},

		// Version
		{
			name:    "version with --version",
			args:    []string{"tucha", "--version"},
			wantCmd: CmdVersion,
		},
		{
			name:    "version with -v",
			args:    []string{"tucha", "-v"},
			wantCmd: CmdVersion,
		},

		// Server commands
		{
			name:    "background",
			args:    []string{"tucha", "--background"},
			wantCmd: CmdBackground,
		},
		{
			name:    "status",
			args:    []string{"tucha", "--status"},
			wantCmd: CmdStatus,
		},
		{
			name:    "stop",
			args:    []string{"tucha", "--stop"},
			wantCmd: CmdStop,
		},
		{
			name:    "config-check",
			args:    []string{"tucha", "--config-check"},
			wantCmd: CmdConfigCheck,
		},

		// Config option
		{
			name:       "custom config path",
			args:       []string{"tucha", "-config", "/etc/tucha/config.yaml"},
			wantCmd:    CmdRun,
			wantConfig: "/etc/tucha/config.yaml",
		},
		{
			name:       "config with background",
			args:       []string{"tucha", "-config", "custom.yaml", "--background"},
			wantCmd:    CmdBackground,
			wantConfig: "custom.yaml",
		},

		// User commands
		{
			name:     "user list",
			args:     []string{"tucha", "--user", "list"},
			wantCmd:  CmdUserList,
			wantArgs: []string{},
		},
		{
			name:     "user list with mask",
			args:     []string{"tucha", "--user", "list", "*@example.com"},
			wantCmd:  CmdUserList,
			wantArgs: []string{"*@example.com"},
		},
		{
			name:     "user add minimal",
			args:     []string{"tucha", "--user", "add", "user@example.com", "password123"},
			wantCmd:  CmdUserAdd,
			wantArgs: []string{"user@example.com", "password123"},
		},
		{
			name:     "user add with quota",
			args:     []string{"tucha", "--user", "add", "user@example.com", "password123", "16GB"},
			wantCmd:  CmdUserAdd,
			wantArgs: []string{"user@example.com", "password123", "16GB"},
		},
		{
			name:     "user remove",
			args:     []string{"tucha", "--user", "remove", "user@example.com"},
			wantCmd:  CmdUserRemove,
			wantArgs: []string{"user@example.com"},
		},
		{
			name:     "user pwd",
			args:     []string{"tucha", "--user", "pwd", "user@example.com", "newpass"},
			wantCmd:  CmdUserPwd,
			wantArgs: []string{"user@example.com", "newpass"},
		},
		{
			name:     "user quota",
			args:     []string{"tucha", "--user", "quota", "user@example.com", "8GB"},
			wantCmd:  CmdUserQuota,
			wantArgs: []string{"user@example.com", "8GB"},
		},
		{
			name:     "user info",
			args:     []string{"tucha", "--user", "info", "user@example.com"},
			wantCmd:  CmdUserInfo,
			wantArgs: []string{"user@example.com"},
		},

		// Errors
		{
			name:    "missing config value",
			args:    []string{"tucha", "-config"},
			wantErr: true,
		},
		{
			name:    "unknown option",
			args:    []string{"tucha", "--unknown"},
			wantErr: true,
		},
		{
			name:    "user without subcommand",
			args:    []string{"tucha", "--user"},
			wantErr: true,
		},
		{
			name:    "user add missing password",
			args:    []string{"tucha", "--user", "add", "user@example.com"},
			wantErr: true,
		},
		{
			name:    "user remove missing email",
			args:    []string{"tucha", "--user", "remove"},
			wantErr: true,
		},
		{
			name:    "user pwd missing password",
			args:    []string{"tucha", "--user", "pwd", "user@example.com"},
			wantErr: true,
		},
		{
			name:    "user quota missing value",
			args:    []string{"tucha", "--user", "quota", "user@example.com"},
			wantErr: true,
		},
		{
			name:    "user info missing email",
			args:    []string{"tucha", "--user", "info"},
			wantErr: true,
		},
		{
			name:    "unknown user subcommand",
			args:    []string{"tucha", "--user", "unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got.Command != tt.wantCmd {
				t.Errorf("Parse() Command = %v, want %v", got.Command, tt.wantCmd)
			}

			if tt.wantConfig != "" && got.ConfigPath != tt.wantConfig {
				t.Errorf("Parse() ConfigPath = %v, want %v", got.ConfigPath, tt.wantConfig)
			}

			if tt.wantArgs != nil {
				if len(got.Args) != len(tt.wantArgs) {
					t.Errorf("Parse() Args length = %v, want %v", len(got.Args), len(tt.wantArgs))
				} else {
					for i, arg := range got.Args {
						if arg != tt.wantArgs[i] {
							t.Errorf("Parse() Args[%d] = %v, want %v", i, arg, tt.wantArgs[i])
						}
					}
				}
			}
		})
	}
}

func TestParseEmptyArgs(t *testing.T) {
	cli, err := Parse([]string{})
	if err != nil {
		t.Errorf("Parse([]) error = %v", err)
	}
	if cli.Command != CmdRun {
		t.Errorf("Parse([]) Command = %v, want CmdRun", cli.Command)
	}
}
