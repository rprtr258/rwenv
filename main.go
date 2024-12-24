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

// CLI flags.
var (
	_inherit   bool
	_verbose   bool
	_envFiles  = cli.NewStringSlice()
	_overrides = cli.NewStringSlice()
	app        = &cli.App{
		Usage: "Run command with environment taken from file",
		UsageText: `Run cmd using env:
	rwenv [-i] [-v] [-e env-file]... [-o VAR=override]... cmd...
Show env to be used:
	rwenv [-i] [-v] [-e env-file]... [-o VAR=override]...

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
)

// Errors.
var (
	errNoEqualSign = errors.New("no equal sign found")
)

// Constants.
const (
	_maxFmtValueLen = 100
	_fmtClipLen     = _maxFmtValueLen/2 - 1
)

type EnvVar struct {
	Name  string
	Value string
}

func readFileLines(envFile string) ([]string, error) {
	content, err := os.ReadFile(envFile)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return strings.Split(string(content), "\n"), nil
}

// splitEnv of form key=value into (key, value, err),
// err if something bad happened during splitting, e.g. no equal sign was found.
func splitEnv(env string) (string, string, error) {
	runes := []rune(env)
	eqSignIdx := strings.IndexRune(env, '=')
	if eqSignIdx == -1 {
		return "", "", errNoEqualSign
	}

	key := string(runes[:eqSignIdx])

	valueRunes := runes[eqSignIdx+1:]
	if len(valueRunes) >= 2 &&
		valueRunes[0] == '"' &&
		valueRunes[len(valueRunes)-1] == '"' {
		valueRunes = valueRunes[1 : len(valueRunes)-1]
	}
	value := string(valueRunes)

	return key, value, nil
}

func collectEnv() (map[string]string, error) {
	if !_inherit && len(_envFiles.Value()) == 0 && len(_overrides.Value()) == 0 {
		return nil, nil
	}

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
			if _verbose {
				log.Printf("    set env  %s=%q\n", key, value)
			}
			envp[key] = value
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
			key, value, err := splitEnv(envVarLine)
			if err != nil {
				if _verbose {
					log.Printf("    ignoring line %q: %s\n", envVarLine, err.Error())
				}

				continue
			}
			if _verbose {
				log.Printf("    set env  %s=%q\n", key, value)
			}
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

func printEnvs(environ map[string]string) {
	if environ == nil {
		envp := os.Environ()
		environ = make(map[string]string, len(envp))
		for _, envVar := range envp {
			parts := strings.SplitN(envVar, "=", 2)
			environ[parts[0]] = parts[1]
		}
	}

	// just print env nicely:
	//   - sorted by name
	//   - cutting long values to begin...end
	//   - with padding separating names and values
	envp := []EnvVar{}
	for k, varValue := range environ {
		if len(varValue) > _maxFmtValueLen {
			varValue = fmt.Sprintf(
				"%s...%s",
				varValue[:_fmtClipLen],
				varValue[len(varValue)-_fmtClipLen:],
			)
		}
		envp = append(envp, EnvVar{
			Name:  k,
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
}

func run(ctx *cli.Context) error {
	args := ctx.Args().Slice()

	envp, err := collectEnv()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		printEnvs(envp)
		return nil
	}

	program, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("look executable path: %w", err)
	}

	if err := syscall.Exec(program, args, envToList(envp)); err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

// TODO: logs colors
func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
