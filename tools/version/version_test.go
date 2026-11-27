package version

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockGitRunner simulates git command outputs for testing.
type mockGitRunner struct {
	outputs map[string]string
	errors  map[string]error
}

func (m *mockGitRunner) Run(args ...string) (string, error) {
	cmd := strings.Join(args, " ")
	if err, exists := m.errors[cmd]; exists {
		return "", err
	}
	if output, exists := m.outputs[cmd]; exists {
		return output, nil
	}

	return "", fmt.Errorf("mock command not found: git %s", cmd)
}

const listTags = "tag -l v*.*.* --sort=-version:refname"

// gitWithTags returns a git mock with the given tags, biggest first, where tag is not found.
func gitWithTags(tags string, notFound ...string) *mockGitRunner {
	m := &mockGitRunner{
		outputs: map[string]string{
			"fetch --all --tags": "",
			listTags:             tags,
		},
		errors: make(map[string]error),
	}
	for _, tag := range notFound {
		m.errors["rev-parse "+tag] = errors.New("tag not found")
	}

	return m
}

type verifyCase struct {
	name      string
	input     string
	mockGit   GitRunner
	expected  string
	expectErr bool
}

func TestVerify(t *testing.T) {
	t.Parallel()

	testCases := []verifyCase{
		{
			name:     "valid patch bump",
			input:    "1.2.4",
			mockGit:  gitWithTags("v1.2.3\nv1.2.2", "v1.2.4"),
			expected: "1.2.4",
		},
		{
			name:     "valid minor bump",
			input:    "1.3.0",
			mockGit:  gitWithTags("v1.2.3\nv1.2.2", "v1.3.0"),
			expected: "1.3.0",
		},
		{
			name:     "valid major bump",
			input:    "2.0.0",
			mockGit:  gitWithTags("v1.2.3\nv1.2.2", "v2.0.0"),
			expected: "2.0.0",
		},
		{
			name:     "v prefix is stripped",
			input:    "v1.3.0",
			mockGit:  gitWithTags("v1.2.3", "v1.3.0"),
			expected: "1.3.0",
		},
		{
			name:     "spaces are trimmed",
			input:    " v1.3.0\n",
			mockGit:  gitWithTags("v1.2.3", "v1.3.0"),
			expected: "1.3.0",
		},
		{
			name:     "first release",
			input:    "0.1.0",
			mockGit:  gitWithTags("", "v0.1.0"),
			expected: "0.1.0",
		},
		{
			name:      "tag already exists",
			input:     "1.2.3",
			mockGit:   &mockGitRunner{outputs: map[string]string{"fetch --all --tags": "", listTags: "v1.2.3", "rev-parse v1.2.3": "commit-hash"}},
			expectErr: true,
		},
		{
			name:      "invalid gap - patch too high",
			input:     "1.2.5",
			mockGit:   gitWithTags("v1.2.3", "v1.2.5"),
			expectErr: true,
		},
		{
			name:      "invalid gap - minor too high",
			input:     "1.4.0",
			mockGit:   gitWithTags("v1.2.3", "v1.4.0"),
			expectErr: true,
		},
		{
			name:      "invalid gap - version regression",
			input:     "1.2.2",
			mockGit:   gitWithTags("v1.2.3", "v1.2.2"),
			expectErr: true,
		},
		{
			name:     "empty bumps the minor of the biggest tag",
			input:    "",
			mockGit:  gitWithTags("v4.55.0\nv4.54.2", "v4.56.0"),
			expected: "4.56.0",
		},
		{
			name:     "empty after a patch release",
			input:    "",
			mockGit:  gitWithTags("v4.55.3\nv4.55.2", "v4.56.0"),
			expected: "4.56.0",
		},
		{
			name:      "empty without tags",
			input:     "",
			mockGit:   gitWithTags(""),
			expectErr: true,
		},
		{
			name:      "git fetch fails",
			input:     "1.2.4",
			mockGit:   &mockGitRunner{errors: map[string]error{"fetch --all --tags": errors.New("network error")}},
			expectErr: true,
		},
		{
			name:      "empty when git fetch fails",
			input:     "",
			mockGit:   &mockGitRunner{errors: map[string]error{"fetch --all --tags": errors.New("network error")}},
			expectErr: true,
		},
		{
			name:  "git tag list fails",
			input: "1.2.4",
			mockGit: &mockGitRunner{
				outputs: map[string]string{"fetch --all --tags": ""},
				errors:  map[string]error{listTags: errors.New("git command failed")},
			},
			expectErr: true,
		},
		{
			name:      "latest tag is not valid semver",
			input:     "1.2.4",
			mockGit:   gitWithTags("v-is-for-invalid", "v1.2.4"),
			expectErr: true,
		},
	}

	// Each would be a valid bump of v4.66.0 under a lenient parser, so only the input check rejects them.
	for name, input := range map[string]string{
		"not a version":  "not-a-version",
		"two v prefixes": "vv4.67.0",
		"upper case V":   "V4.67.0",
		"no patch":       "4.67",
		"major only":     "4",
		"leading zero":   "04.67.0",
		"pre-release":    "4.67.0-rc1",
		"build metadata": "4.67.0+build",
		"trailing text":  "4.67.0 final",
		"four parts":     "4.67.0.1",
	} {
		testCases = append(testCases, verifyCase{
			name:      "invalid: " + name,
			input:     input,
			mockGit:   gitWithTags("v4.66.0", "v4.67.0"),
			expectErr: true,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Verify(tc.mockGit, tc.input)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}
