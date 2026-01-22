package service

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func ControlContainer(containerName string, action string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	switch action {
	case "start":
		return cli.ContainerStart(ctx, containerName, container.StartOptions{})
	case "stop":
		return cli.ContainerStop(ctx, containerName, container.StopOptions{})
	case "restart":
		return cli.ContainerRestart(ctx, containerName, container.StopOptions{})
	}

	return nil
}
