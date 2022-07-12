package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/urfave/cli/v2"
)

var (
	_inherit   bool
	_verbose   bool
	_envFiles  = cli.NewStringSlice()
	_overrides = cli.NewStringSlice()
	_envVarRE  = regexp.MustCompile("^([A-Z0-9_]+)=(.*)$")
)

func readFileLines(envFile string) ([]string, error) {
	content, err := os.ReadFile(envFile)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n"), nil
}

func collectEnv() (map[string]string, error) {
	envp := make(map[string]string)
	if _inherit {
		if _verbose {
			log.Println("inheriting env vars...")
		}
		for _, envVarLine := range os.Environ() {
			match := _envVarRE.FindStringSubmatch(envVarLine)
			if _verbose {
				log.Printf("setting env var %s\n", envVarLine)
			}
			envp[match[1]] = match[2]
		}
	}
	for _, envFile := range _envFiles.Value() {
		if _verbose {
			log.Println("reading env file", envFile)
		}
		lines, err := readFileLines(envFile)
		if err != nil {
			return nil, err
		}
		for _, envVarLine := range lines {
			if _envVarRE.MatchString(envVarLine) {
				if _verbose {
					log.Printf("    set env  %q\n", envVarLine)
				}
				match := _envVarRE.FindStringSubmatch(envVarLine)
				envp[match[1]] = match[2]
			}
		}
	}
	for _, envVarLine := range _overrides.Value() {
		if !_envVarRE.MatchString(envVarLine) {
			return nil, fmt.Errorf("wrong env var format: %q", envVarLine)
		}
		if _verbose {
			log.Printf("override %q\n", envVarLine)
		}
		match := _envVarRE.FindStringSubmatch(envVarLine)
		envp[match[1]] = match[2]
	}
	return envp, nil
}

func envToList(env map[string]string) []string {
	envp := make([]string, 0, len(env))
	for varName, varValue := range env {
		envp = append(envp, fmt.Sprintf("%s=%s", varName, varValue))
	}
	return envp
}

func run(ctx *cli.Context) error {
	args := ctx.Args().Slice()

	if len(args) == 0 {
		// TODO: maybe pretty print env in that case
		return fmt.Errorf("command to run is not provided")
	}

	program, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}

	envp, err := collectEnv()
	if err != nil {
		return err
	}

	return syscall.Exec(program, args, envToList(envp))
}

func main() {
	app := &cli.App{
		Usage: "Run command with environment taken from file",
		UsageText: `rwenv [-i] [-v] [-e env-file]... [-o override]... cmd...

Example:
	rwenv -e .env env`,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "env",
				Aliases:     []string{"e"},
				Usage:       "env files to take vars from",
				Destination: _envFiles,
			},
			&cli.StringSliceFlag{
				Name:        "override",
				Aliases:     []string{"o"},
				Usage:       "additional env vars in form of VAR_NAME=VALUE",
				Destination: _overrides,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Usage:       "print var reading info",
				Destination: &_verbose,
			},
			&cli.BoolFlag{
				Name:        "inherit",
				Aliases:     []string{"i"},
				Usage:       "inherit shell env vars",
				Destination: &_inherit,
			},
		},
		Action:                 run,
		HideHelpCommand:        true,
		UseShortOptionHandling: true,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
