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
	"time"
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
				Category:    "",
				DefaultText: "",
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				HasBeenSet:  false,
				Value:       nil,
				EnvVars:     nil,
				TakesFile:   false,
				Action:      nil,
				KeepSpace:   false,
			},
			&cli.StringSliceFlag{
				Name:        "override",
				Aliases:     []string{"o"},
				Usage:       "additional env vars in form of VAR_NAME=VALUE",
				Destination: _overrides,
				Category:    "",
				DefaultText: "",
				FilePath:    "",
				Required:    false,
				Hidden:      false,
				HasBeenSet:  false,
				Value:       nil,
				EnvVars:     nil,
				TakesFile:   false,
				Action:      nil,
				KeepSpace:   false,
			},
			&cli.BoolFlag{
				Name:               "verbose",
				Aliases:            []string{"v"},
				Usage:              "print var reading info",
				Destination:        &_verbose,
				Category:           "",
				DefaultText:        "",
				FilePath:           "",
				Required:           false,
				Hidden:             false,
				HasBeenSet:         false,
				Value:              false,
				EnvVars:            nil,
				Action:             nil,
				Count:              nil,
				DisableDefaultText: false,
			},
			&cli.BoolFlag{
				Name:               "inherit",
				Aliases:            []string{"i"},
				Usage:              "inherit shell env vars",
				Destination:        &_inherit,
				Category:           "",
				DefaultText:        "",
				FilePath:           "",
				Required:           false,
				Hidden:             false,
				HasBeenSet:         false,
				Value:              false,
				EnvVars:            nil,
				Action:             nil,
				Count:              nil,
				DisableDefaultText: false,
			},
		},
		Action:                    run,
		HideHelpCommand:           true,
		UseShortOptionHandling:    true,
		Name:                      "",
		HelpName:                  "",
		ArgsUsage:                 "",
		Version:                   "",
		Description:               "",
		DefaultCommand:            "",
		Commands:                  nil,
		EnableBashCompletion:      false,
		HideHelp:                  false,
		HideVersion:               false,
		BashComplete:              nil,
		Before:                    nil,
		After:                     nil,
		CommandNotFound:           nil,
		OnUsageError:              nil,
		InvalidFlagAccessHandler:  nil,
		Compiled:                  time.Time{},
		Authors:                   nil,
		Copyright:                 "",
		Reader:                    nil,
		Writer:                    nil,
		ErrWriter:                 nil,
		ExitErrHandler:            nil,
		Metadata:                  nil,
		ExtraInfo:                 nil,
		CustomAppHelpTemplate:     "",
		SliceFlagSeparator:        "",
		DisableSliceFlagSeparator: false,
		Suggest:                   false,
		AllowExtFlags:             false,
		SkipFlagParsing:           false,
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

func printEnvs() {
	// just print env nicely:
	//   - sorted by name
	//   - cutting long values to begin...end
	//   - with padding separating names and values
	envp := []EnvVar{}
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2) // TODO: change to regex
		varValue := parts[1]
		if len(varValue) > _maxFmtValueLen {
			varValue = fmt.Sprintf(
				"%s...%s",
				varValue[:_fmtClipLen],
				varValue[len(varValue)-_fmtClipLen:],
			)
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
}

func run(ctx *cli.Context) error {
	args := ctx.Args().Slice()

	if len(args) == 0 {
		printEnvs()
		return nil
	}

	program, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("look executable path: %w", err)
	}

	envp, err := collectEnv()
	if err != nil {
		return err
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
