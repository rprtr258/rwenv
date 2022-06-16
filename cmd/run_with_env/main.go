package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var envVarLine *regexp.Regexp

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
	envFile, err := cmd.Flags().GetString("env")
	if err != nil {
		return err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}
	parent, err := cmd.Flags().GetBool("parent")
	if err != nil {
		return err
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
				fmt.Println("set env", line)
			}
		}
	}
	proc := exec.Command(args[0], args[1:]...)
	if parent {
		proc.Env = os.Environ()
	}
	proc.Env = append(proc.Env, envp...)
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
	}()
	go func() {
		io.Copy(os.Stdout, stdout)
	}()
	go func() {
		io.Copy(os.Stderr, stderr)
	}()
	return proc.Wait()
}

func main() {
	rootCmd := cobra.Command{
		Use:     "rwenv",
		Short:   "Run command with environment taken from file",
		Args:    cobra.MinimumNArgs(1),
		RunE:    run,
		Example: "TODO: add",
	}
	rootCmd.Flags().StringP("env", "e", "", "Env file to take vars from")
	rootCmd.Flags().BoolP("verbose", "v", false, "Print steps")
	rootCmd.Flags().BoolP("parent", "p", false, "Add parent env vars")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
