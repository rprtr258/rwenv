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
		Example: "TODO: add",
	}
)

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

func run(cmd *cobra.Command, args []string) error {
	proc := exec.Command(args[0], args[1:]...)
	if inherit {
		if verbose {
			log.Println("inheriting env vars...")
		}
		proc.Env = os.Environ()
	}
	for _, envFile := range envFiles {
		if verbose {
			log.Println("reading env file", envFile)
		}
		lines, err := readFileLines(envFile)
		if err != nil {
			return err
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
		proc.Env = append(proc.Env, envp...)
	}
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
	proc.Start()
	go func() {
		io.Copy(stdin, os.Stdin)
		os.Stdin.Close()
	}()
	go func() {
		io.Copy(os.Stdout, stdout)
		os.Stdout.Close()
	}()
	go func() {
		io.Copy(os.Stderr, stderr)
	}()
	return proc.Wait()
}

func main() {
	rootCmd.Flags().StringSliceVarP(&envFiles, "env", "e", nil, "Env file to take vars from")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Print steps")
	rootCmd.Flags().BoolVarP(&inherit, "inherit", "i", false, "Inherit shell env vars")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
