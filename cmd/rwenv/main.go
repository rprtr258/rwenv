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

const (
	setEnvLineFormat  = "    set env  %q\n"
	overrideVarFormat = "override %q\n"
)

var (
	envVarLine *regexp.Regexp

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

func init() {
	var err error
	envVarLine, err = regexp.Compile("^[A-Z_]+=.*$")
	if err != nil {
		log.Fatal(err.Error())
	}
}

func readFileLines(envFile string) ([]string, error) {
	content, err := os.ReadFile(envFile)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n"), nil
}

func makeEnvList(opts Options) ([]string, error) {
	var res []string
	if opts.inherit {
		if opts.verbose {
			log.Println("inheriting env vars...")
		}
		res = os.Environ()
	}
	for _, envFile := range opts.envFiles {
		if opts.verbose {
			log.Println("reading env file", envFile)
		}
		lines, err := readFileLines(envFile)
		if err != nil {
			return nil, err
		}
		envp := []string{}
		for _, line := range lines {
			if envVarLine.MatchString(line) {
				envp = append(envp, line)
				if opts.verbose {
					log.Printf(setEnvLineFormat, line)
				}
			}

		}
		res = append(res, envp...)
	}
	for _, envVar := range opts.envOverrides {
		if !envVarLine.MatchString(envVar) {
			return nil, fmt.Errorf("wrong env var format: %q", envVar)
		}
		if opts.verbose {
			log.Printf(overrideVarFormat, envVar)
		}
		res = append(res, envVar)
	}
	return res, nil
}

func run(opts Options) error {
	if opts.help {
		fmt.Println(usage)
		return nil
	}
	envp, err := makeEnvList(opts)
	if err != nil {
		return err
	}
	program, err := exec.LookPath(opts.cmd[0])
	if err != nil {
		return err
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
