package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

const Binary = "temp/ehugo"

var Default = Run

func Run(ctx context.Context) (err error) {
	return run(ctx, tempDir, buildService, runService)
}

func run(ctx context.Context, task ...func(c context.Context) error) (err error) {
	for _, t := range task {
		if err = t(ctx); err != nil {
			return err
		}
	}
	return nil
}

func tempDir(_ context.Context) error {
	return os.MkdirAll("temp", 0755)
}

func buildService(_ context.Context) error {
	cmd := exec.Command("go", "build", "-o", Binary, "service/main/main.go")
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return err
	}
	return nil
}

func runService(_ context.Context) error {
	cmd := exec.Command("sudo", "ADDRESS=10.0.0.82:80", Binary)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
