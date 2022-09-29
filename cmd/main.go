package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/urfave/cli/v2"
)

var (
	_inherit   bool
	_verbose   bool
	_envFiles  = cli.NewStringSlice()
	_overrides = cli.NewStringSlice()
)

type EnvVar struct {
	Name  string
	Value string
}

func readFileLines(envFile string) ([]string, error) {
	content, err := os.ReadFile(envFile)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n"), nil
}

// splitEnv of form key=value into (key, value, err),
// err if something bad happened during splitting, e.g. no equal sign was found
func splitEnv(env string) (string, string, error) {
	runes := []rune(env)
	eqSignIdx := strings.IndexRune(env, '=')
	if eqSignIdx == -1 {
		return "", "", errors.New("No equal sign found")
	}

	return string(runes[:eqSignIdx]), string(runes[eqSignIdx+1:]), nil
}

func collectEnv() (map[string]string, error) {
	envp := make(map[string]string)
	if _inherit {
		if _verbose {
			log.Println("inheriting env vars...")
		}
		for _, envVarLine := range os.Environ() {
			key, value, err := splitEnv(envVarLine)
			if err != nil {
				// TODO: paint red
				log.Printf("incorrect env var %q: %s\n", envVarLine, err.Error())
				continue
			}
			log.Printf("    set env  %s=%q\n", key, value)
			envp[key] = value
		}
	}
	log.Println("files", _envFiles.Value())
	for _, envFile := range _envFiles.Value() {
		if _verbose {
			log.Println("reading env file", envFile)
		}
		lines, err := readFileLines(envFile)
		if err != nil {
			return nil, err
		}
		for _, envVarLine := range lines {
			key, value, err := splitEnv(envVarLine)
			if err != nil {
				if _verbose {
					log.Printf("    ignoring line %q: %s\n", envVarLine, err.Error())
				}
				continue
			}
			log.Printf("    set env  %s=%q\n", key, value)
			envp[key] = value
		}
	}
	for _, envVarLine := range _overrides.Value() {
		key, value, err := splitEnv(envVarLine)
		if err != nil {
			log.Printf("    ignoring override %q: %s\n", envVarLine, err.Error())
		} else if _verbose {
			log.Printf("override %q\n", envVarLine)
		}
		envp[key] = value
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
		// just print env nicely:
		//   - sorted by name
		//   - cutting long values to begin...end
		//   - with padding separating names and values
		envp := []EnvVar{}
		for _, envVar := range os.Environ() {
			parts := strings.SplitN(envVar, "=", 2)
			varValue := parts[1]
			if len(varValue) > 100 {
				varValue = varValue[:49] + "..." + varValue[len(varValue)-49:]
			}
			envp = append(envp, EnvVar{
				Name:  parts[0],
				Value: varValue,
			})
		}
		sort.Slice(envp, func(i, j int) bool {
			return envp[i].Name < envp[j].Name
		})
		maxLen := 0
		for _, envVar := range envp {
			if len(envVar.Name) > maxLen {
				maxLen = len(envVar.Name)
			}
		}
		pad := strings.Repeat(" ", maxLen)
		for _, envVar := range envp {
			padding := pad[:maxLen-utf8.RuneCountInString(envVar.Name)]
			fmt.Printf("%s%s = %q\n", envVar.Name, padding, envVar.Value)
		}
		return nil
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

// TODO: logs colors
func main() {
	app := &cli.App{
		Usage: "Run command with environment taken from file",
		UsageText: `Run cmd using env:
	rwenv [-i] [-v] [-e env-file]... [-o VAR=override]... cmd...
Show env to be used:
	rwenv [-i] [-v] [-e env-file]... [-o VAR=override]... cmd...

Example:
	rwenv             # show env
	rwenv -e .env env # show env with added vars from .env`,
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
