package sysinfo

import (
	"context"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/tsaikd/gosys2bigquery/applog"
	"golang.org/x/sync/errgroup"
)

type sysInfoDockerContainer struct {
	Timestamp time.Time
	Hostname  string
	ID        string
	Name      string
	CPUStats  struct {
		CPUUsage struct {
			TotalUsage bqInteger
		}
	}
	MemoryStats struct {
		Usage bqInteger
	}
	Network struct {
		RxBytes bqInteger
		TxBytes bqInteger
	}
}

var dockerClient *docker.Client

func newDockerClient(ctx context.Context, endpoint string) (err error) {
	if dockerClient, err = docker.NewClient(endpoint); err != nil {
		return
	}
	if err = dockerClient.PingWithContext(ctx); err != nil {
		return
	}
	return
}

func getDockerStats(
	ctx context.Context,
	interval time.Duration,
	statsChan chan<- sysInfoDockerContainer,
	endpoint string,
	keepNamePrefixSlash bool,
) (err error) {
	if dockerClient == nil {
		if err = newDockerClient(ctx, endpoint); err != nil {
			return
		}
	}

	go func() {
		applog.Trace(startDockerGetContainersStatsLoop(ctx, interval, statsChan, keepNamePrefixSlash))
	}()

	return nil
}

func startDockerGetContainersStatsLoop(
	ctxx context.Context,
	interval time.Duration,
	msgChan chan<- sysInfoDockerContainer,
	keepNamePrefixSlash bool,
) (err error) {
	eg, ctx := errgroup.WithContext(ctxx)

	// listen running containers
	eg.Go(func() error {
		containers, err := dockerClient.ListContainers(docker.ListContainersOptions{
			Context: ctx,
		})
		if err != nil {
			return err
		}
		for _, container := range containers {
			func(containerID string, containerName string) {
				eg.Go(func() error {
					return startDockerGetContainerStatsLoop(ctx, containerID, containerName, interval, msgChan)
				})
			}(container.ID, filterContainerName(container.Names[0], keepNamePrefixSlash))
		}
		return nil
	})

	// listen for running in future containers
	eg.Go(func() error {
		dockerEventChan := make(chan *docker.APIEvents, 100)
		if err := dockerClient.AddEventListener(dockerEventChan); err != nil {
			return err
		}

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case dockerEvent, ok := <-dockerEventChan:
				if !ok {
					return nil
				}
				if dockerEvent.Status == "start" {
					container, err := dockerClient.InspectContainer(dockerEvent.ID)
					if err != nil {
						applog.Trace(err)
						continue
					}

					func(containerID string, containerName string) {
						eg.Go(func() error {
							return startDockerGetContainerStatsLoop(ctx, containerID, containerName, interval, msgChan)
						})
					}(dockerEvent.ID, filterContainerName(container.Name, keepNamePrefixSlash))
				}
			}
		}
	})

	return eg.Wait()
}

func startDockerGetContainerStatsLoop(
	ctxx context.Context,
	containerID string,
	containerName string,
	interval time.Duration,
	msgChan chan<- sysInfoDockerContainer,
) (err error) {
	eg, ctx := errgroup.WithContext(ctxx)
	retry := 3
	statsChan := make(chan *docker.Stats, 5)
	for err == nil || retry > 0 {
		eg.Go(func() error {
			nextReceive := time.Time{}
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case stats, ok := <-statsChan:
					if !ok {
						return nil
					}
					if stats.Read.Before(nextReceive) {
						continue
					}
					nextReceive = stats.Read.Add(interval - 200*time.Millisecond)

					msg := sysInfoDockerContainer{
						Timestamp: stats.Read,
						Hostname:  hostname,
						ID:        containerID,
						Name:      containerName,
					}
					msg.CPUStats.CPUUsage.TotalUsage = bqInteger(stats.CPUStats.CPUUsage.TotalUsage)
					msg.MemoryStats.Usage = bqInteger(stats.MemoryStats.Usage)
					msg.Network.RxBytes = bqInteger(stats.Network.RxBytes)
					msg.Network.TxBytes = bqInteger(stats.Network.TxBytes)
					msgChan <- msg
				}
			}
		})

		err = dockerClient.Stats(docker.StatsOptions{
			ID:      containerID,
			Stats:   statsChan,
			Stream:  true,
			Context: ctx,
		})
		if err != nil && strings.Contains(err.Error(), "connection refused") {
			retry--
			time.Sleep(500 * time.Millisecond)
			continue
		}
		break
	}

	return eg.Wait()
}

func filterContainerName(containerName string, keepNamePrefixSlash bool) string {
	if !keepNamePrefixSlash {
		return strings.TrimPrefix(containerName, "/")
	}
	return containerName
}
