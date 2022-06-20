package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

var (
	envVarRE = regexp.MustCompile("^([A-Z0-9_]+)=(.*)$")

	usage = `rwenv [flags | [-e env-file]... | [-o override]...] cmd...
Run command with environment taken from file

Example:
  rwenv -e .env env

Flags:
  -e, --env env-file       env files to take vars from
  -h, --help               show this message
  -i, --inherit            inherit shell env vars
  -o, --override override  additional env vars in form of VAR_NAME=VALUE
  -v, --verbose            print var reading info`
)

type Options struct {
	envFiles     []string
	envOverrides []string
	verbose      bool
	inherit      bool
	help         bool
	cmd          []string
}

func readFileLines(envFile string) ([]string, error) {
	content, err := os.ReadFile(envFile)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n"), nil
}

func makeEnvList(opts Options) (map[string]string, error) {
	res := make(map[string]string)
	if opts.inherit {
		if opts.verbose {
			log.Println("inheriting env vars...")
		}
		for _, envVarLine := range os.Environ() {
			match := envVarRE.FindStringSubmatch(envVarLine)
			if opts.verbose {
				log.Printf("setting env var %s\n", envVarLine)
			}
			res[match[1]] = match[2]
		}
	}
	for _, envFile := range opts.envFiles {
		if opts.verbose {
			log.Println("reading env file", envFile)
		}
		lines, err := readFileLines(envFile)
		if err != nil {
			return nil, err
		}
		for _, envVarLine := range lines {
			if envVarRE.MatchString(envVarLine) {
				if opts.verbose {
					log.Printf("    set env  %q\n", envVarLine)
				}
				match := envVarRE.FindStringSubmatch(envVarLine)
				res[match[1]] = match[2]
			}

		}
	}
	for _, envVarLine := range opts.envOverrides {
		if !envVarRE.MatchString(envVarLine) {
			return nil, fmt.Errorf("wrong env var format: %q", envVarLine)
		}
		if opts.verbose {
			log.Printf("override %q\n", envVarLine)
		}
		match := envVarRE.FindStringSubmatch(envVarLine)
		res[match[1]] = match[2]
	}
	return res, nil
}

func run(opts Options) error {
	if opts.help {
		fmt.Println(usage)
		return nil
	}
	env, err := makeEnvList(opts)
	if err != nil {
		return err
	}
	program, err := exec.LookPath(opts.cmd[0])
	if err != nil {
		return err
	}
	envp := make([]string, 0, len(env))
	for varName, varValue := range env {
		envp = append(envp, fmt.Sprintf("%s=%s", varName, varValue))
	}
	return syscall.Exec(program, opts.cmd, envp)
}

func parseArgs() (opts Options, err error) {
	argv := os.Args
	argN := len(argv)
	for i := 1; i < argN; i++ {
		switch {
		case argv[i] == "-e" || argv[i] == "--env":
			i++
			if i == argN {
				err = fmt.Errorf("env file is expected after %s", argv[i-1])
				return
			}
			opts.envFiles = append(opts.envFiles, argv[i])
		case argv[i] == "-o" || argv[i] == "--override":
			i++
			if i == argN {
				err = fmt.Errorf("env var in form of VAR_NAME=VALUE is expected after %s", argv[i-1])
				return
			}
			opts.envOverrides = append(opts.envOverrides, argv[i])
		case argv[i] == "-v" || argv[i] == "--verbose":
			opts.verbose = true
		case argv[i] == "-i" || argv[i] == "--inherit":
			opts.inherit = true
		case argv[i] == "-h" || argv[i] == "--help":
			opts.help = true
			return
		default:
			opts.cmd = argv[i:]
			return
		}
	}
	err = fmt.Errorf("command to run is not provided")
	return
}

func main() {
	opts, err := parseArgs()
	if err != nil {
		log.Fatal(err.Error())
	}
	if err := run(opts); err != nil {
		log.Fatal(err.Error())
	}
}
