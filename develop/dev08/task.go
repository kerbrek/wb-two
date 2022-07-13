package main

/*
=== Взаимодействие с ОС ===

Необходимо реализовать собственный шелл

встроенные команды: cd/pwd/echo/kill/ps
поддержать fork/exec команды
конвеер на пайпах

Реализовать утилиту netcat (nc) клиент
принимать данные из stdin и отправлять в соединение (tcp/udp)
Программа должна проходить все тесты. Код должен проходить проверки go vet и golint.
*/

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"github.com/mattn/go-shellwords"
)

func main() {
	opts := new(options)
	reader := os.Stdin
	writer := os.Stdout
	errs := os.Stderr
	if status := do(reader, writer, errs, opts); status.exit {
		if status.err != nil {
			fmt.Fprintln(errs, status.err)
		}
		os.Exit(status.code)
	} else {
		fmt.Fprintln(errs, "Status.exit must be true")
		os.Exit(1)
	}
}

type options struct {
	username    string
	hostname    string
	homeDir     string
	workDir     string
	pid         int
	lastCmdCode int
	cmdParser   *shellwords.Parser
}

func (opts *options) complete() error {
	user, err := user.Current()
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	opts.username = user.Username
	opts.homeDir = user.HomeDir
	opts.hostname = hostname
	opts.workDir = workDir
	opts.pid = os.Getpid()

	opts.cmdParser = shellwords.NewParser()
	opts.cmdParser.ParseEnv = true

	return nil
}

func (opts *options) prompt() string {
	dir := strings.Replace(opts.workDir, opts.homeDir, "~", 1)
	prompt := fmt.Sprintf("%s@%s:%s $ ", opts.username, opts.hostname, dir)
	return prompt
}

var ErrInvalidSyntax = errors.New("syntax error")

func do(in io.Reader, out io.Writer, errs io.Writer, opts *options) Status {
	if err := opts.complete(); err != nil {
		return Status{exit: true, code: 1, err: err}
	}

	reader := bufio.NewReader(in)

	for {
		if _, pErr := fmt.Fprint(out, opts.prompt()); pErr != nil {
			return Status{exit: true, code: 1, err: pErr}
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Fprintln(out, "")
				return Status{exit: true, code: 0, err: nil}
			}
			return Status{exit: true, code: 1, err: err}
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		cmds, err := parsePipeline(input, opts)
		if err != nil {
			if _, pErr := fmt.Fprintln(errs, err); pErr != nil {
				return Status{exit: true, code: 1, err: pErr}
			}
			continue
		}

		pipeline := &Pipeline{
			cmds:   cmds,
			stdin:  in,
			stdout: out,
			stderr: errs,
		}

		if status := pipeline.exec(opts); status.exit {
			return status
		}
	}
}

type Status struct {
	exit bool
	code int
	err  error
}

func parsePipeline(input string, opts *options) ([]*CMD, error) {
	if strings.HasPrefix(input, "|") || strings.HasSuffix(input, "|") {
		return nil, ErrInvalidSyntax
	}

	const (
		unquoted = iota
		singleQuoted
		doubleQuoted
		escaped
	)
	parts := []string{}
	b := strings.Builder{}
	state := unquoted

	for _, r := range input {
		switch state {
		case unquoted:
			if r == '|' {
				parts = append(parts, strings.TrimSpace(b.String()))
				b.Reset()
			} else if r == '\'' {
				state = singleQuoted
				b.WriteRune(r)
			} else if r == '"' {
				state = doubleQuoted
				b.WriteRune(r)
			} else {
				b.WriteRune(r)
			}
		case singleQuoted:
			if r == '\'' {
				state = unquoted
			}
			b.WriteRune(r)
		case doubleQuoted:
			if r == '"' {
				state = unquoted
			} else if r == '\\' {
				state = escaped
			}
			b.WriteRune(r)
		case escaped:
			state = doubleQuoted
			b.WriteRune(r)
		}
	}

	parts = append(parts, strings.TrimSpace(b.String()))

	if state != unquoted {
		return nil, ErrInvalidSyntax
	}

	cmds := make([]*CMD, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return nil, ErrInvalidSyntax
		}

		cmdParts, err := opts.cmdParser.Parse(part)
		if err != nil {
			return nil, err
		}

		cmd := &CMD{
			prog:     cmdParts[0],
			args:     cmdParts[1:],
			pipeline: true,
			rawInput: part,
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

type Pipeline struct {
	cmds   []*CMD
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (pl *Pipeline) exec(opts *options) Status {
	if len(pl.cmds) == 1 {
		cmd := pl.cmds[0]
		cmd.pipeline = false
		cmd.stdin = pl.stdin
		cmd.stdout = pl.stdout
		cmd.stderr = pl.stderr
		return cmd.exec(opts)
	}

	var status Status
	var reader *os.File
	var writer *os.File

	for i, cmd := range pl.cmds {
		cmd.stderr = pl.stderr

		r, w, err := os.Pipe()
		if err != nil {
			return Status{exit: true, code: 1, err: err}
		}

		if i == 0 {
			cmd.stdin = pl.stdin
			cmd.stdout = w
			reader = r
			writer = w
		} else if i == len(pl.cmds)-1 {
			cmd.stdin = reader
			cmd.stdout = pl.stdout
			writer = w
		} else {
			cmd.stdin = reader
			cmd.stdout = w
			reader = r
			writer = w
		}

		status = cmd.exec(opts)
		writer.Close()
	}

	return status
}

type CMD struct {
	prog     string
	args     []string
	pipeline bool
	rawInput string
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
}

var ErrUnsupportedCommand = errors.New("unsupported command")

func (cmd *CMD) exec(opts *options) Status {
	var status Status

	switch cmd.prog {
	case "exit":
		status = exit(cmd, opts)
	case "cd":
		status = cd(cmd, opts)
	case "pwd":
		status = pwd(cmd, opts)
	case "echo":
		status = echo(cmd, opts)
	case "kill":
		status = kill(cmd, opts)
	case "ps":
		status = run(cmd)
	case "fork":
		status = fork(cmd, opts)
	case "exec":
		status = execute(cmd, opts)
	default:
		// status = Status{code: 1, err: ErrUnsupportedCommand}
		status = run(cmd)
	}

	if status.exit {
		return status
	}

	if status.err != nil {
		if _, pErr := fmt.Fprintln(cmd.stderr, status.err); pErr != nil {
			return Status{exit: true, code: 1, err: pErr}
		}
	}

	opts.lastCmdCode = status.code
	return status
}

func run(cmd *CMD) Status {
	c := exec.Command(cmd.prog, cmd.args...)
	c.Stdin = cmd.stdin
	c.Stdout = cmd.stdout
	c.Stderr = cmd.stderr
	err := c.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return Status{code: exitErr.ExitCode()}
		}

		return Status{code: 1, err: err}
	}

	return Status{code: 0}
}

var ErrExitTooManyArgs = errors.New("exit: too many arguments")

func exit(cmd *CMD, opts *options) Status {
	if len(cmd.args) > 1 {
		return Status{code: 1, err: ErrExitTooManyArgs}
	}

	if len(cmd.args) == 0 {
		status := Status{exit: true, code: opts.lastCmdCode}
		if cmd.pipeline {
			status.exit = false
		}
		return status
	}

	code, err := strconv.ParseUint(cmd.args[0], 10, 8)
	if err != nil {
		return Status{code: 1, err: err}
	}

	status := Status{exit: true, code: int(code)}
	if cmd.pipeline {
		status.exit = false
	}
	return status
}

func cd(cmd *CMD, opts *options) Status {
	var dir string
	if len(cmd.args) == 0 {
		dir = opts.homeDir
	} else {
		dir = strings.Join(cmd.args, " ")
	}

	if strings.HasPrefix(dir, "~") {
		dir = strings.Replace(dir, "~", opts.homeDir, 1)
	}

	if err := os.Chdir(dir); err != nil {
		return Status{code: 1, err: err}
	}

	workDir, err := os.Getwd()
	if err != nil {
		return Status{code: 1, err: err}
	}

	opts.workDir = workDir
	return Status{code: 0}
}

func pwd(cmd *CMD, opts *options) Status {
	workDir, err := os.Getwd()
	if err != nil {
		return Status{code: 1, err: err}
	}

	if _, pErr := fmt.Fprintln(cmd.stdout, workDir); pErr != nil {
		return Status{exit: true, code: 1, err: pErr}
	}

	return Status{code: 0}
}

func echo(cmd *CMD, opts *options) Status {
	mapping := func(key string) string {
		if key == "?" {
			return fmt.Sprint(opts.lastCmdCode)
		}
		return os.Getenv(key)
	}

	parts, _ := shellwords.Parse(cmd.rawInput)
	args := parts[1:]
	for i, arg := range args {
		args[i] = os.Expand(arg, mapping)
	}

	if _, pErr := fmt.Fprintln(cmd.stdout, strings.Join(args, " ")); pErr != nil {
		return Status{exit: true, code: 1, err: pErr}
	}

	return Status{code: 0}
}

var ErrKillPidRequired = errors.New("usage: kill pid")

func kill(cmd *CMD, opts *options) Status {
	if len(cmd.args) == 0 {
		return Status{code: 1, err: ErrKillPidRequired}
	}

	pid, err := strconv.ParseInt(cmd.args[0], 10, 32)
	if err != nil {
		return Status{code: 1, err: err}
	}

	self := int64(opts.pid)
	if pid == self {
		return Status{code: 0}
	}

	process, _ := os.FindProcess(int(pid))
	if err := process.Kill(); err != nil {
		return Status{code: 1, err: err}
	}

	return Status{code: 0}
}

func fork(cmd *CMD, opts *options) Status {
	if len(cmd.args) == 0 {
		return Status{code: 0}
	}

	cmd.prog = cmd.args[0]
	cmd.args = cmd.args[1:]
	c := exec.Command(cmd.prog, cmd.args...)
	if err := c.Start(); err != nil {
		return Status{code: 1, err: err}
	}

	return Status{code: 0}
}

func execute(cmd *CMD, opts *options) Status {
	if len(cmd.args) == 0 {
		return Status{code: 0}
	}

	cmd.prog = cmd.args[0]
	cmd.args = cmd.args[1:]
	status := run(cmd)
	if status.err != nil {
		return status
	}

	if cmd.pipeline {
		return status
	}

	status.exit = true
	return status
}
