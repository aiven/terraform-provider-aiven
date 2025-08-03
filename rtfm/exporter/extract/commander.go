package extract

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type CommandOptions struct {
	// WorkDir is the directory where the command will be executed
	WorkDir string

	// Env is a list of environment variables in the form "KEY=value"
	// If nil, the current process's environment will be used
	Env []string

	// StdIn is the input to the command
	StdIn io.Reader

	// StdOut captures the command's output
	StdOut io.Writer

	// StdErr captures the command's error output
	// If nil, stderr will be captured internally
	StdErr io.Writer

	// SkipErrorCheck disables automatic error checking after command execution
	SkipErrorCheck bool
}

type Commander interface {
	// Execute runs the specified command with the given arguments and options
	Execute(command string, args []string, opts *CommandOptions) error
}

type TerraformCommander struct {
	BinaryPath string
}

// NewTerraformCommander creates a new TerraformCommander
func NewTerraformCommander(binaryPath string) *TerraformCommander {
	return &TerraformCommander{
		BinaryPath: binaryPath,
	}
}

// Execute runs a terraform command with the given arguments and options
func (t *TerraformCommander) Execute(command string, args []string, opts *CommandOptions) error {
	if opts == nil {
		opts = &CommandOptions{}
	}

	cmdArgs := append([]string{command}, args...)
	cmd := exec.Command(t.BinaryPath, cmdArgs...)

	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	if opts.Env != nil {
		cmd.Env = opts.Env
	} else {
		cmd.Env = os.Environ()
	}

	if opts.StdIn != nil {
		cmd.Stdin = opts.StdIn
	}

	if opts.StdOut != nil {
		cmd.Stdout = opts.StdOut
	}

	var stderr bytes.Buffer
	if opts.StdErr != nil {
		cmd.Stderr = opts.StdErr
	} else {
		cmd.Stderr = &stderr
	}

	err := cmd.Run()

	// check for errors unless disabled
	if err != nil && !opts.SkipErrorCheck {
		if cmd.Stderr == &stderr {
			return fmt.Errorf("command '%s %s' failed: %w, stderr: %s",
				t.BinaryPath,
				formatArgs(cmdArgs),
				err,
				stderr.String())
		}

		return fmt.Errorf("command '%s %s' failed: %w",
			t.BinaryPath,
			formatArgs(cmdArgs),
			err)
	}

	return nil
}

func formatArgs(args []string) string {
	var result bytes.Buffer
	for i, arg := range args {
		if i > 0 {
			result.WriteString(" ")
		}
		result.WriteString(arg)
	}

	return result.String()
}
