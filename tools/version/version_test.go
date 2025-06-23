package version

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestVerify(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		newVersion string
		mockGit    GitRunner
		expectErr  bool
	}{
		{
			name:       "valid patch bump",
			newVersion: "1.2.4",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "v1.2.3\nv1.2.2",
				},
				errors: map[string]error{
					"rev-parse v1.2.4": errors.New("tag not found"),
				},
			},
			expectErr: false,
		},
		{
			name:       "valid minor bump",
			newVersion: "1.3.0",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "v1.2.3\nv1.2.2",
				},
				errors: map[string]error{
					"rev-parse v1.3.0": errors.New("tag not found"),
				},
			},
			expectErr: false,
		},
		{
			name:       "valid major bump",
			newVersion: "2.0.0",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "v1.2.3\nv1.2.2",
				},
				errors: map[string]error{
					"rev-parse v2.0.0": errors.New("tag not found"),
				},
			},
			expectErr: false,
		},
		{
			name:       "tag already exists",
			newVersion: "1.2.3",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags": "",
					"rev-parse v1.2.3":   "commit-hash",
				},
			},
			expectErr: true,
		},
		{
			name:       "invalid gap - patch too high",
			newVersion: "1.2.5",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "v1.2.3",
				},
				errors: map[string]error{
					"rev-parse v1.2.5": errors.New("tag not found"),
				},
			},
			expectErr: true,
		},
		{
			name:       "invalid format",
			newVersion: "not-a-version",
			mockGit:    &mockGitRunner{},
			expectErr:  true,
		},
		{
			name:       "first release",
			newVersion: "0.1.0",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "",
				},
				errors: map[string]error{
					"rev-parse v0.1.0": errors.New("tag not found"),
				},
			},
			expectErr: false,
		},
		{
			name:       "valid patch bump",
			newVersion: "1.2.4",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "v1.2.3\nv1.2.2",
				},
				errors: map[string]error{
					"rev-parse v1.2.4": errors.New("tag not found"),
				},
			},
			expectErr: false,
		},
		{
			name:       "valid minor bump",
			newVersion: "1.3.0",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "v1.2.3\nv1.2.2",
				},
				errors: map[string]error{
					"rev-parse v1.3.0": errors.New("tag not found"),
				},
			},
			expectErr: false,
		},
		{
			name:       "valid major bump",
			newVersion: "2.0.0",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "v1.2.3\nv1.2.2",
				},
				errors: map[string]error{
					"rev-parse v2.0.0": errors.New("tag not found"),
				},
			},
			expectErr: false,
		},
		{
			name:       "tag already exists",
			newVersion: "1.2.3",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags": "",
					"rev-parse v1.2.3":   "commit-hash", // command succeeds, meaning tag exists
				},
			},
			expectErr: true,
		},
		{
			name:       "invalid gap - patch too high",
			newVersion: "1.2.5",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "v1.2.3",
				},
				errors: map[string]error{
					"rev-parse v1.2.5": errors.New("tag not found"),
				},
			},
			expectErr: true,
		},
		{
			name:       "invalid format for new version",
			newVersion: "not-a-version",
			mockGit:    &mockGitRunner{},
			expectErr:  true,
		},
		{
			name:       "first release",
			newVersion: "0.1.0",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "",
				},
				errors: map[string]error{
					"rev-parse v0.1.0": errors.New("tag not found"),
				},
			},
			expectErr: false,
		},
		{
			name:       "invalid gap - minor too high",
			newVersion: "1.4.0",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "v1.2.3",
				},
				errors: map[string]error{
					"rev-parse v1.4.0": errors.New("tag not found"),
				},
			},
			expectErr: true,
		},
		{
			name:       "invalid gap - version regression",
			newVersion: "1.2.2",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "v1.2.3",
				},
				errors: map[string]error{
					"rev-parse v1.2.2": errors.New("tag not found"), // It shouldn't get this far
				},
			},
			expectErr: true,
		},
		{
			name:       "pre-release version is invalid bump",
			newVersion: "1.2.4-rc1",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "v1.2.3",
				},
				errors: map[string]error{
					"rev-parse v1.2.4-rc1": errors.New("tag not found"),
				},
			},
			expectErr: true,
		},
		{
			name:       "git fetch fails",
			newVersion: "1.2.4",
			mockGit: &mockGitRunner{
				errors: map[string]error{
					"fetch --all --tags": errors.New("network error"),
				},
			},
			expectErr: true,
		},
		{
			name:       "git tag list fails",
			newVersion: "1.2.4",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags": "",
				},
				errors: map[string]error{
					"rev-parse v1.2.4":                      errors.New("tag not found"),
					"tag -l v*.*.* --sort=-version:refname": errors.New("git command failed"),
				},
			},
			expectErr: true,
		},
		{
			name:       "latest tag is not valid semver",
			newVersion: "1.2.4",
			mockGit: &mockGitRunner{
				outputs: map[string]string{
					"fetch --all --tags":                    "",
					"tag -l v*.*.* --sort=-version:refname": "v-is-for-invalid",
				},
				errors: map[string]error{
					"rev-parse v1.2.4": errors.New("tag not found"),
				},
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := Verify(tc.mockGit, tc.newVersion)
			assert.Equal(t, tc.expectErr, err != nil, "expected error: %v, got: %v", tc.expectErr, err)
		})
	}
}
