package main

import (
	"context"
)

// Ci is the Dagger CI pipeline module for the petstore API.
type Ci struct{}

// Build compiles the Go server binary.
func (m *Ci) Build(ctx context.Context, source *Directory) (string, error) {
	return goBase(source).
		WithExec([]string{"go", "build", "-o", "bin/server", "./cmd/server"}).
		Stdout(ctx)
}

// Lint runs go vet across all packages.
func (m *Ci) Lint(ctx context.Context, source *Directory) (string, error) {
	return goBase(source).
		WithExec([]string{"go", "vet", "./..."}).
		Stdout(ctx)
}

// Test runs the full test suite. Requires Docker socket access for testcontainers.
func (m *Ci) Test(ctx context.Context, source *Directory) (string, error) {
	return goBase(source).
		WithUnixSocket("/var/run/docker.sock", dag.Host().UnixSocket("/var/run/docker.sock")).
		WithEnvVariable("DOCKER_HOST", "unix:///var/run/docker.sock").
		WithEnvVariable("TESTCONTAINERS_RYUK_DISABLED", "true").
		WithExec([]string{"go", "test", "./...", "-timeout", "120s"}).
		Stdout(ctx)
}

func goBase(source *Directory) *Container {
	return dag.Container().
		From("golang:1.24-alpine").
		WithMountedCache("/root/go/pkg/mod", dag.CacheVolume("gomod-cache")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("gobuild-cache")).
		WithDirectory("/src", source).
		WithWorkdir("/src")
}
