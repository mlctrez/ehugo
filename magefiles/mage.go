package main

import (
	"context"
	"os"
	"os/exec"
)

const Binary = "temp/ehugo"

var Default = BuildRun

func Deploy(ctx context.Context) (err error) {
	if err = tasks(ctx, tempDir, buildService); err != nil {
		return err
	}
	if err = runStd(exec.Command("scp", Binary, "ehugo:/tmp/ehugo")); err != nil {
		return err
	}
	return runStd(exec.Command("ssh", "ehugo", "/tmp/ehugo", "-action", "deploy"))
}

func BuildRun(ctx context.Context) (err error) {
	return tasks(ctx, tempDir, buildService, runService)
}

func tempDir(_ context.Context) error {
	return os.MkdirAll("temp", 0755)
}

func buildService(_ context.Context) error {
	return runStd(exec.Command("go", "build", "-o", Binary, "service/main/main.go"))
}

func runService(_ context.Context) error {
	return runStd(exec.Command("sudo", "ADDRESS=10.0.0.82:80", Binary))
}

func tasks(ctx context.Context, task ...func(c context.Context) error) (err error) {
	for _, t := range task {
		if err = t(ctx); err != nil {
			return err
		}
	}
	return nil
}

func runStd(cmd *exec.Cmd) error {
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
