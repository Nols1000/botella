package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/agonzalezro/ava/bot"
	"github.com/agonzalezro/ava/config"
	"github.com/agonzalezro/ava/plugin"
)

func init() {
	if os.Getenv("DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
}

func inferConfigPath() (string, error) {
	paths := []string{"ava.yml", "ava.yaml"}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("No %s file found!\n", strings.Join(paths, " or "))
}

func loadPlugins(config *config.Config) ([]*plugin.Plugin, error) {
	var plugins []*plugin.Plugin
	for _, pluginConfig := range config.Plugins {
		plugin, err := plugin.New(
			pluginConfig.Image,
			pluginConfig.Environment,
		)
		if err != nil {
			return nil, fmt.Errorf("Error loading plugin (image: %s): %v", pluginConfig.Image, err)
		}

		// TODO: this is a little bit ugly
		plugin.RunOnlyOnChannels = pluginConfig.OnlyChannels
		plugin.RunOnlyOnDirectMessages = pluginConfig.OnlyDirectMessages
		plugin.RunOnlyOnMentions = pluginConfig.OnlyMentions

		log.Infof("Plugin (%s) loaded.", pluginConfig.Image)
		plugins = append(plugins, plugin)
	}
	return plugins, nil
}

func loadAdapters(config *config.Config) ([]bot.Adapter, error) {
	var adapters []bot.Adapter
	for _, adapterConfig := range config.Adapters {
		adapter, err := bot.New(adapterConfig.Name, adapterConfig.Environment)
		if err != nil {
			return nil, fmt.Errorf("Error loading adapter (%s): %v", adapterConfig.Name, err)
		}

		log.Infof("Adapter (%s) loaded.", adapterConfig.Name)
		adapters = append(adapters, adapter)
	}
	return adapters, nil
}

func listenAndReply(adapters []bot.Adapter, plugins []*plugin.Plugin) {
	var wg sync.WaitGroup
	signalsCh := make(chan os.Signal, 1)
	signal.Notify(signalsCh, os.Interrupt)

	for _, adapter := range adapters {
		wg.Add(1)

		stdinCh, stdoutCh, stderrCh := adapter.RunAndAttach()
		go func(adapter bot.Adapter, stdinCh, stdoutCh chan bot.Message, stderrCh chan error) {
			for {
				select {
				case m := <-stdinCh:
					log.Debugf("Message received: %+v", m)
					for _, p := range plugins {
						if !adapter.ShouldRun(p, &m) {
							log.Debugf("Not running plugin (%s) for: %+v", p.Image, m)
							continue
						}
						log.Debugf("Running plugin (%s) for: %+v", p.Image, m)

						pluginResponse, err := p.Run(m.Body)
						if err != nil {
							stderrCh <- err
							continue
						}
						pluginResponse = strings.TrimSuffix(pluginResponse, "\n")

						log.Debugf("Plugin (%s) response: %s", p.Image, pluginResponse)
						stdoutCh <- bot.Message{Channel: m.Channel, Body: pluginResponse}
					}
				case err := <-stderrCh:
					log.Error(err)
				case <-signalsCh:
					for i := 0; i < len(adapters); i++ {
						wg.Done()
					}
				}
			}
		}(adapter, stdinCh, stdoutCh, stderrCh)
	}

	wg.Wait()

	log.Info("Teardown...")
	for _, plugin := range plugins {
		plugin.Stop()
	}
}

func main() {
	configPath, err := inferConfigPath()
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	config, err := config.NewFromFile(configPath)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	adapters, err := loadAdapters(config)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	plugins, err := loadPlugins(config)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	listenAndReply(adapters, plugins)
}
