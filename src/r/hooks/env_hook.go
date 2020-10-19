package hooks

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/paketo-buildpacks/packit/pexec"
)

const (
	EnvStart = "<<<<ENVIRONMENT_START>>>>"
	EnvStop  = "<<<<ENVIRONMENT_STOP>>>>"
)

func init() {
	_, err := os.Stat("/tmp/app/r.env.sh")
	if err == nil {
		libbuildpack.AddHook(EnvHook{})
	}
}

type EnvHook struct {
	libbuildpack.DefaultHook
}

func (h EnvHook) BeforeCompile(compiler *libbuildpack.Stager) error {
	compiler.Logger().BeginStep("Setting up R environment using r.env.sh")

	commands := []string{
		fmt.Sprintf("source %s", filepath.Join(compiler.BuildDir(), "r.env.sh")),
		fmt.Sprintf("echo %q", EnvStart),
		"env",
		fmt.Sprintf("echo %q", EnvStop),
	}
	buffer := bytes.NewBuffer(nil)

	bash := pexec.NewExecutable("bash")
	err := bash.Execute(pexec.Execution{
		Args:   []string{"-c", strings.Join(commands, " && ")},
		Stdout: buffer,
		Stderr: buffer,
	})
	if err != nil {
		return fmt.Errorf("%s\n%w\n", buffer, err)
	}

	var parseEnv bool
	scanner := bufio.NewScanner(buffer)
	for scanner.Scan() {
		line := scanner.Text()

		switch line {
		case EnvStart:
			parseEnv = true
		case EnvStop:
			parseEnv = false
		default:
			if parseEnv {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					os.Setenv(parts[0], parts[1])
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
