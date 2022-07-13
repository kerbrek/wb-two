package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestShell(t *testing.T) {
	t.Run(("test shell"), func(t *testing.T) {
		opts := new(options)
		inbuf := &bytes.Buffer{}
		outbuf := &bytes.Buffer{}

		cmd := exec.Command("sleep", "7")
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}

		inbuf.WriteString("cd testdata\n")
		inbuf.WriteString("pwd\n")
		inbuf.WriteString("echo Pipe | exit 1 | echo Hello World\n")
		inbuf.WriteString(fmt.Sprintf("kill %d\n", cmd.Process.Pid))
		inbuf.WriteString("cd ..\n")
		inbuf.WriteString("exit\n")

		status := do(inbuf, outbuf, outbuf, opts)
		expectedStatus := Status{exit: true, code: 0, err: nil}
		if expectedStatus != status {
			fmt.Println(expectedStatus)
			fmt.Println("--")
			fmt.Println(status)
			t.Fatal("Not equal")
		}

		data, err := os.ReadFile("testdata/output.txt")
		if err != nil {
			t.Fatal(err)
		}

		expectedOutput := string(data)

		prefix := strings.TrimSuffix(opts.prompt(), " $ ")
		cwd, _ := os.Getwd()
		output := outbuf.String()
		cleaned := strings.ReplaceAll(output, prefix, "")
		cleaned = strings.ReplaceAll(cleaned, cwd, "")

		if expectedOutput != cleaned {
			fmt.Println(expectedOutput)
			fmt.Println("--")
			fmt.Println(cleaned)
			t.Fatal("Not equal")
		}

		if cmd.ProcessState.ExitCode() != -1 {
			t.Fatal("Not equal")
		}
	})
}
