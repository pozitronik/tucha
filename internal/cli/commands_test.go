package cli

import (
	"strings"
	"testing"
)

func TestCommand_Constants(t *testing.T) {
	// Verify that command constants are unique and in expected order
	commands := []struct {
		cmd  Command
		name string
	}{
		{CmdRun, "CmdRun"},
		{CmdHelp, "CmdHelp"},
		{CmdVersion, "CmdVersion"},
		{CmdBackground, "CmdBackground"},
		{CmdStatus, "CmdStatus"},
		{CmdStop, "CmdStop"},
		{CmdConfigCheck, "CmdConfigCheck"},
		{CmdUserList, "CmdUserList"},
		{CmdUserAdd, "CmdUserAdd"},
		{CmdUserRemove, "CmdUserRemove"},
		{CmdUserPwd, "CmdUserPwd"},
		{CmdUserQuota, "CmdUserQuota"},
		{CmdUserInfo, "CmdUserInfo"},
	}

	seen := make(map[Command]string)
	for _, tc := range commands {
		if existing, ok := seen[tc.cmd]; ok {
			t.Errorf("Duplicate command value: %s and %s have same value %d", existing, tc.name, tc.cmd)
		}
		seen[tc.cmd] = tc.name
	}

	// Verify CmdRun is the default (iota starts at 0)
	if CmdRun != 0 {
		t.Errorf("CmdRun = %d, want 0 (should be default)", CmdRun)
	}
}

func TestExitCode_Constants(t *testing.T) {
	// Verify exit codes are as documented
	tests := []struct {
		name     string
		got      int
		expected int
	}{
		{"ExitSuccess", ExitSuccess, 0},
		{"ExitError", ExitError, 1},
		{"ExitConfigError", ExitConfigError, 2},
		{"ExitAlreadyRunning", ExitAlreadyRunning, 3},
		{"ExitNotRunning", ExitNotRunning, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestHelpText(t *testing.T) {
	help := HelpText()

	t.Run("contains program name", func(t *testing.T) {
		if !strings.Contains(help, "Tucha") {
			t.Error("HelpText() should contain 'Tucha'")
		}
	})

	t.Run("contains usage section", func(t *testing.T) {
		if !strings.Contains(help, "Usage:") {
			t.Error("HelpText() should contain 'Usage:'")
		}
	})

	t.Run("contains options section", func(t *testing.T) {
		if !strings.Contains(help, "Options:") {
			t.Error("HelpText() should contain 'Options:'")
		}
	})

	t.Run("contains commands section", func(t *testing.T) {
		if !strings.Contains(help, "Commands:") {
			t.Error("HelpText() should contain 'Commands:'")
		}
	})

	t.Run("contains user management section", func(t *testing.T) {
		if !strings.Contains(help, "User Management:") {
			t.Error("HelpText() should contain 'User Management:'")
		}
	})

	t.Run("contains examples section", func(t *testing.T) {
		if !strings.Contains(help, "Examples:") {
			t.Error("HelpText() should contain 'Examples:'")
		}
	})

	// Verify all documented commands are present
	requiredCommands := []string{
		"--help",
		"--version",
		"--background",
		"--status",
		"--stop",
		"--config-check",
		"--user list",
		"--user add",
		"--user remove",
		"--user pwd",
		"--user quota",
		"--user info",
	}

	for _, cmd := range requiredCommands {
		t.Run("contains "+cmd, func(t *testing.T) {
			if !strings.Contains(help, cmd) {
				t.Errorf("HelpText() should contain %q", cmd)
			}
		})
	}

	// Verify config option is documented
	t.Run("contains config option", func(t *testing.T) {
		if !strings.Contains(help, "-config") {
			t.Error("HelpText() should contain '-config'")
		}
	})

	// Verify quota format examples
	t.Run("contains quota format examples", func(t *testing.T) {
		if !strings.Contains(help, "8GB") && !strings.Contains(help, "16GB") {
			t.Error("HelpText() should contain quota format examples like '8GB' or '16GB'")
		}
	})

	// Verify help text ends properly (no trailing garbage)
	t.Run("ends with newline", func(t *testing.T) {
		if !strings.HasSuffix(help, "\n") {
			t.Error("HelpText() should end with newline")
		}
	})
}

func TestHelpText_NotEmpty(t *testing.T) {
	help := HelpText()
	if len(help) < 100 {
		t.Errorf("HelpText() length = %d, want at least 100 characters", len(help))
	}
}
