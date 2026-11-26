package version

import (
	"bytes"
	"errors"
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

// Verify returns the release version as `x.y.z`, after checking it is a valid and sequential bump of the
// biggest git tag. The input must be a plain `x.y.z`, with or without one `v`.
// An empty input resolves to the next minor of the biggest tag.
func Verify(git GitRunner, input string) (string, error) {
	input = strings.TrimSpace(input)
	slog.Info("Starting version verification", slog.String("version", input))

	var newVersion *semver.Version
	if input != "" {
		v, err := semver.StrictNewVersion(strings.TrimPrefix(input, "v"))
		if err != nil {
			return "", fmt.Errorf("invalid release version %q, want x.y.z or vx.y.z: %w", input, err)
		}

		if v.Prerelease() != "" || v.Metadata() != "" {
			return "", fmt.Errorf("invalid release version %q: pre-release and build metadata are not allowed", input)
		}

		newVersion = v
	}

	// fetch tags from remote
	if _, err := git.Run("fetch", "--all", "--tags"); err != nil {
		return "", fmt.Errorf("failed to fetch git tags: %w", err)
	}

	latestTagStr, latestVersion, err := latestTag(git)
	if err != nil {
		return "", err
	}

	if newVersion == nil {
		if latestVersion == nil {
			return "", errors.New("no version given and no git tags to bump: pass the version")
		}

		next := latestVersion.IncMinor()
		newVersion = &next
		slog.Info("No version given", slog.String("latest", latestTagStr), slog.String("next", next.String()))
	}

	newVersionTag := "v" + newVersion.String()
	if _, err := git.Run("rev-parse", newVersionTag); err == nil {
		return "", fmt.Errorf("tag %s already exists", newVersionTag)
	}

	if latestVersion == nil {
		// first tag, no previous versions
		return newVersion.String(), nil
	}

	slog.Info("Comparing with the latest tag", slog.String("latest", latestTagStr), slog.String("new", newVersionTag))

	expectedMajor := latestVersion.IncMajor()
	expectedMinor := latestVersion.IncMinor()
	expectedPatch := latestVersion.IncPatch()

	if newVersion.Equal(&expectedMajor) || newVersion.Equal(&expectedMinor) || newVersion.Equal(&expectedPatch) {
		slog.Info("✅ Version bump is valid")
		return newVersion.String(), nil
	}

	return "", fmt.Errorf(
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

// latestTag returns the biggest `v*.*.*` git tag, or nil if there are none.
func latestTag(git GitRunner) (string, *semver.Version, error) {
	out, err := git.Run("tag", "-l", "v*.*.*", "--sort=-version:refname")
	if err != nil {
		return "", nil, fmt.Errorf("failed to list git tags: %w", err)
	}

	tag, _, _ := strings.Cut(out, "\n")
	if tag == "" {
		return "", nil, nil
	}

	v, err := semver.NewVersion(strings.TrimPrefix(tag, "v"))
	if err != nil {
		// this should ideally not happen if tags follow semver
		return "", nil, fmt.Errorf("could not parse latest tag %q: %w", tag, err)
	}

	return tag, v, nil
}

// Validate checks if a string is a valid semantic version
func Validate(versionStr string) error {
	_, err := semver.NewVersion(versionStr)
	if err != nil {
		return fmt.Errorf("%q is not a valid semantic version: %w", versionStr, err)
	}

	return nil
}
