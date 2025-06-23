package version

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// GitRunner defines an interface for running Git commands.
type GitRunner interface {
	Run(args ...string) (string, error)
}

// RealGitRunner is the production implementation that runs actual git commands.
type RealGitRunner struct{}

func (r RealGitRunner) Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git command failed: %w\n%s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// Verify checks if a new version is a valid and sequential semantic version bump.
func Verify(git GitRunner, newVersionStr string) error {
	slog.Info("Starting version verification", slog.String("version", newVersionStr))

	newVersion, err := semver.NewVersion(newVersionStr)
	if err != nil {
		return fmt.Errorf("invalid version format for %q: %w", newVersionStr, err)
	}

	// fetch tags from remote
	if _, err := git.Run("fetch", "--all", "--tags"); err != nil {
		return fmt.Errorf("failed to fetch git tags: %w", err)
	}

	newVersionTag := "v" + newVersion.String()
	if _, err := git.Run("rev-parse", newVersionTag); err == nil {
		return fmt.Errorf("tag %s already exists", newVersionTag)
	}

	// find the latest existing semver tag
	latestTagsOutput, err := git.Run("tag", "-l", "v*.*.*", "--sort=-version:refname")
	if err != nil {
		return fmt.Errorf("failed to list git tags: %w", err)
	}

	tags := strings.Split(latestTagsOutput, "\n")
	if len(tags) == 0 || tags[0] == "" {
		// first tag, no previous versions
		return nil
	}
	latestTagStr := tags[0]
	fmt.Printf("🔍 Latest tag is %s. Comparing with %s.\n", latestTagStr, newVersionTag)

	latestVersion, err := semver.NewVersion(strings.TrimPrefix(latestTagStr, "v"))
	if err != nil {
		// this should ideally not happen if tags follow semver
		return fmt.Errorf("could not parse latest tag %q: %w", latestTagStr, err)
	}

	expectedMajor := latestVersion.IncMajor()
	expectedMinor := latestVersion.IncMinor()
	expectedPatch := latestVersion.IncPatch()

	if newVersion.Equal(&expectedMajor) || newVersion.Equal(&expectedMinor) || newVersion.Equal(&expectedPatch) {
		slog.Info("✅ Version bump is valid")
		return nil
	}

	return fmt.Errorf(
		"invalid version gap detected.\n"+
			"  New version '%s' is not a sequential increment from latest '%s'.\n"+
			"  Valid next versions would be: v%s, v%s, or v%s",
		newVersionTag,
		latestTagStr,
		expectedMajor.String(),
		expectedMinor.String(),
		expectedPatch.String(),
	)
}

// Validate checks if a string is a valid semantic version
func Validate(versionStr string) error {
	_, err := semver.NewVersion(versionStr)
	if err != nil {
		return fmt.Errorf("%q is not a valid semantic version: %w", versionStr, err)
	}

	return nil
}
