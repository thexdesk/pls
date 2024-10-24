package dindcmd

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/hinshun/pls/docker/dind"
	"github.com/hinshun/pls/docker/dockercli"
	"github.com/palantir/stacktrace"

	"gopkg.in/urfave/cli.v2"
)

func CreateDind(c *cli.Context) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return stacktrace.Propagate(err, "failed to create docker client from env: %s", err)
	}

	spec := dind.DindSpec{
		Name:                  c.String("name"),
		Image:                 c.String("image"),
		MITMProxyName:         c.String("mitm"),
		RegistryServerAddress: c.String("registry"),
		RegistryUsername:      c.String("username"),
		RegistryPassword:      c.String("password"),
	}

	err = dockercli.LazyImageLoad(ctx, cli, spec.Image)
	if err != nil {
		return stacktrace.Propagate(err, "failed to load dind image")
	}

	_, err = dind.New(ctx, cli, spec)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create new dind")
	}

	return nil
}
