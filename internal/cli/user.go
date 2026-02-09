package cli

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/pozitronik/tucha/internal/application/service"
	"github.com/pozitronik/tucha/internal/domain/repository"
)

// UserCommands handles CLI user management operations.
type UserCommands struct {
	userService *service.UserService
	userRepo    repository.UserRepository
}

// NewUserCommands creates a new UserCommands instance.
func NewUserCommands(userService *service.UserService, userRepo repository.UserRepository) *UserCommands {
	return &UserCommands{
		userService: userService,
		userRepo:    userRepo,
	}
}

// List outputs all users, optionally filtered by email pattern.
// Pattern supports * as wildcard (e.g., "*@example.com").
func (c *UserCommands) List(w io.Writer, pattern string) error {
	users, err := c.userService.ListWithUsage()
	if err != nil {
		return fmt.Errorf("listing users: %w", err)
	}

	// Filter by pattern if provided
	if pattern != "" {
		filtered := make([]service.UserWithUsage, 0)
		for _, u := range users {
			if matchPattern(u.Email, pattern) {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}

	if len(users) == 0 {
		fmt.Fprintln(w, "No users found")
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Email\tPassword\tQuota\tUsed\tAdmin\tSizeLimit\tHistory")

	for _, u := range users {
		quota := FormatByteSize(u.QuotaBytes)
		used := FormatByteSize(u.BytesUsed)
		admin := ""
		if u.IsAdmin {
			admin = "yes"
		}
		sizeLimit := "unlimited"
		if u.FileSizeLimit > 0 {
			sizeLimit = FormatByteSize(u.FileSizeLimit)
		}
		history := "free"
		if u.VersionHistory {
			history = "paid"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", u.Email, u.Password, quota, used, admin, sizeLimit, history)
	}

	return tw.Flush()
}

// Add creates a new user with the given parameters.
func (c *UserCommands) Add(w io.Writer, email, password string, quotaStr string) error {
	var quotaBytes int64
	if quotaStr != "" {
		var err error
		quotaBytes, err = ParseByteSize(quotaStr)
		if err != nil {
			return fmt.Errorf("invalid quota %q: %w", quotaStr, err)
		}
	}

	user, err := c.userService.Create(email, password, false, quotaBytes)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	fmt.Fprintf(w, "User created: %s (quota: %s)\n", user.Email, FormatByteSize(user.QuotaBytes))
	return nil
}

// Remove deletes a user by email.
func (c *UserCommands) Remove(w io.Writer, email string) error {
	user, err := c.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", email)
	}

	if err := c.userService.Delete(user.ID); err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}

	fmt.Fprintf(w, "User removed: %s\n", email)
	return nil
}

// SetPassword updates a user's password.
func (c *UserCommands) SetPassword(w io.Writer, email, password string) error {
	user, err := c.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", email)
	}

	user.Password = password
	if err := c.userService.Update(user); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	fmt.Fprintf(w, "Password updated for: %s\n", email)
	return nil
}

// SetQuota updates a user's storage quota.
func (c *UserCommands) SetQuota(w io.Writer, email, quotaStr string) error {
	quotaBytes, err := ParseByteSize(quotaStr)
	if err != nil {
		return fmt.Errorf("invalid quota %q: %w", quotaStr, err)
	}

	user, err := c.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", email)
	}

	user.QuotaBytes = quotaBytes
	if err := c.userService.Update(user); err != nil {
		return fmt.Errorf("updating quota: %w", err)
	}

	fmt.Fprintf(w, "Quota updated for %s: %s\n", email, FormatByteSize(quotaBytes))
	return nil
}

// SetSizeLimit updates a user's file size limit.
func (c *UserCommands) SetSizeLimit(w io.Writer, email, sizeStr string) error {
	sizeBytes, err := ParseByteSize(sizeStr)
	if err != nil {
		return fmt.Errorf("invalid size %q: %w", sizeStr, err)
	}

	user, err := c.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", email)
	}

	user.FileSizeLimit = sizeBytes
	if err := c.userService.Update(user); err != nil {
		return fmt.Errorf("updating file size limit: %w", err)
	}

	if sizeBytes == 0 {
		fmt.Fprintf(w, "File size limit for %s: unlimited\n", email)
	} else {
		fmt.Fprintf(w, "File size limit for %s: %s\n", email, FormatByteSize(sizeBytes))
	}
	return nil
}

// SetHistory updates a user's version history setting.
func (c *UserCommands) SetHistory(w io.Writer, email, mode string) error {
	var enabled bool
	switch strings.ToLower(mode) {
	case "on", "true", "1", "paid":
		enabled = true
	case "off", "false", "0", "free":
		enabled = false
	default:
		return fmt.Errorf("invalid mode %q: use on/off", mode)
	}

	user, err := c.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", email)
	}

	user.VersionHistory = enabled
	if err := c.userService.Update(user); err != nil {
		return fmt.Errorf("updating version history: %w", err)
	}

	label := "free"
	if enabled {
		label = "paid"
	}
	fmt.Fprintf(w, "Version history for %s: %s\n", email, label)
	return nil
}

// Info displays detailed information about a user.
func (c *UserCommands) Info(w io.Writer, email string) error {
	user, err := c.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", email)
	}

	// Get usage info
	users, err := c.userService.ListWithUsage()
	if err != nil {
		return fmt.Errorf("getting usage: %w", err)
	}

	var used int64
	for _, u := range users {
		if u.ID == user.ID {
			used = u.BytesUsed
			break
		}
	}

	fmt.Fprintf(w, "Email:          %s\n", user.Email)
	fmt.Fprintf(w, "Password:       %s\n", user.Password)
	fmt.Fprintf(w, "Admin:          %v\n", user.IsAdmin)
	fmt.Fprintf(w, "Quota:          %s\n", FormatByteSize(user.QuotaBytes))
	fmt.Fprintf(w, "Used:           %s\n", FormatByteSize(used))

	if user.QuotaBytes > 0 {
		pct := float64(used) / float64(user.QuotaBytes) * 100
		fmt.Fprintf(w, "Usage:          %.1f%%\n", pct)
	}

	if user.FileSizeLimit > 0 {
		fmt.Fprintf(w, "File size limit: %s\n", FormatByteSize(user.FileSizeLimit))
	} else {
		fmt.Fprintf(w, "File size limit: unlimited\n")
	}

	if user.VersionHistory {
		fmt.Fprintf(w, "Version history: paid\n")
	} else {
		fmt.Fprintf(w, "Version history: free\n")
	}

	return nil
}

// matchPattern performs simple wildcard matching.
// Supports * as a wildcard that matches any characters.
func matchPattern(s, pattern string) bool {
	// Convert pattern to lowercase for case-insensitive matching
	s = strings.ToLower(s)
	pattern = strings.ToLower(pattern)

	// Simple wildcard matching
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		// No wildcards, exact match
		return s == pattern
	}

	// Check prefix (part before first *)
	if parts[0] != "" && !strings.HasPrefix(s, parts[0]) {
		return false
	}

	// Check suffix (part after last *)
	if last := parts[len(parts)-1]; last != "" && !strings.HasSuffix(s, last) {
		return false
	}

	// Check middle parts
	pos := len(parts[0])
	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(s[pos:], parts[i])
		if idx < 0 {
			return false
		}
		pos += idx + len(parts[i])
	}

	return true
}
