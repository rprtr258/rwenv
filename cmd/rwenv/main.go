package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

const (
	setEnvLineFormat = "    set env  %q\n"
	ignoreLineFormat = "    ignoring %q\n"
)

var (
	envVarLine *regexp.Regexp

	envFiles []string
	verbose  bool
	inherit  bool
	rootCmd  = cobra.Command{
		Use:     "rwenv",
		Short:   "Run command with environment taken from file",
		Args:    cobra.MinimumNArgs(1),
		RunE:    run,
		Example: "rwenv -e .env env",
	}
)

func init() {
	rootCmd.Flags().StringSliceVarP(&envFiles, "env", "e", nil, "Env files to take vars from")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Print steps")
	rootCmd.Flags().BoolVarP(&inherit, "inherit", "i", false, "Inherit shell env vars")

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

func makeEnvList() ([]string, error) {
	var res []string
	if inherit {
		if verbose {
			log.Println("inheriting env vars...")
		}
		res = os.Environ()
	}
	for _, envFile := range envFiles {
		if verbose {
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
				if verbose {
					log.Printf(setEnvLineFormat, line)
				}
			} else if verbose {
				log.Printf(ignoreLineFormat, line)
			}

		}
		copy(res, envp)
	}
	return res, nil
}

func run(cmd *cobra.Command, args []string) error {
	proc := exec.Command(args[0], args[1:]...)
	envp, err := makeEnvList()
	if err != nil {
		return err
	}
	proc.Env = envp
	stdin, err := proc.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := proc.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := proc.StderrPipe()
	if err != nil {
		return err
	}
	if err := proc.Start(); err != nil {
		return err
	}
	go func() {
		if _, err := io.Copy(stdin, os.Stdin); err != nil {
			log.Fatal(err.Error())
		}
		os.Stdin.Close()
	}()
	go func() {
		if _, err := io.Copy(os.Stdout, stdout); err != nil {
			log.Fatal(err.Error())
		}
		os.Stdout.Close()
	}()
	go func() {
		if _, err := io.Copy(os.Stderr, stderr); err != nil {
			log.Fatal(err.Error())
		}
	}()
	return proc.Wait()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
