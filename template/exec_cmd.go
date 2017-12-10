package template

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

func RemoveTrailingNewlines(execCmd ExecCmdFunc) ExecCmdFunc {
	return func(command, dir string, env []string) (string, error) {
		s, err := execCmd(command, dir, env)
		if err != nil {
			return "", err
		}
		return strings.TrimRight(s, "\n"), nil
	}
}

// ExecCmd executes a command and returns all of its stdout as a string.
// Stderr is suppressed and ignored. Stdin is empty.
// If the command fails then the error message contains the parameters
// of the command execution along with the mixed stdout+stderr of the execution.
//
// This function can be used as a parameter to RemoveTrailingNewlines or as
// a possible value of ExecuteOptions.ExecCmd.
func ExecCmd(command, dir string, env []string) (string, error) {
	var name string
	var carg string
	if runtime.GOOS == "windows" {
		carg = "/C"
		name = os.Getenv("COMSPEC")
		if name == "" {
			name = "cmd.exe"
		}
	} else {
		carg = "-c"
		name = "bash"
		_, err := exec.LookPath(name)
		if err != nil {
			name = "sh"
			_, err = exec.LookPath(name)
			if err != nil {
				return "", err
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	var stdOut bytes.Buffer
	var all bytes.Buffer
	syncedAll := &syncWriter{Writer: &all}

	cmd := exec.CommandContext(ctx, name, carg, command)
	cmd.Env = env
	cmd.Dir = dir
	cmd.Stderr = syncedAll
	cmd.Stdout = io.MultiWriter(&stdOut, syncedAll)

	err := cmd.Run()
	if err != nil {
		errMsg := strings.Join([]string{
			"command execution error: " + err.Error(),
			"command: " + command,
			"directory: " + dir,
			"env vars: " + fmt.Sprintf("%q", env),
			"----- BEGIN OUTPUT -----",
			all.String(),
		}, "\n") + "----- END OUTPUT -----"
		return "", errors.New(errMsg)
	}

	return stdOut.String(), nil
}

type syncWriter struct {
	mu sync.Mutex
	io.Writer
}

func (o *syncWriter) Write(p []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.Writer.Write(p)
}
