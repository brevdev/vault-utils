package main

import (
	"flag"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type config struct {
	SystemdService string
	ConfigFilePath string
	LogLevel       log.Level
}

func main() {
	c, err := getConfig()
	if err != nil {
		panic(err)
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
	log.SetLevel(c.LogLevel)

	err = run(c)
	if err != nil {
		panic(err)
	}
}

func getConfig() (*config, error) {
	systemdService := flag.String("service", "", "The systemd service to restart")
	configFilePath := flag.String("configPath", "", "Path to config file to watch")
	logLevel := flag.String("logLevel", "info", "Ex: trace, debug, info, warn, error")
	flag.Parse()

	if *systemdService == "" {
		return nil, fmt.Errorf("service must be provided")
	}

	if *configFilePath == "" {
		return nil, fmt.Errorf("configPath must be provided")
	}

	l, err := log.ParseLevel(*logLevel)
	if err != nil {
		return nil, wrapAndTrace(err)
	}

	return &config{SystemdService: *systemdService, ConfigFilePath: *configFilePath, LogLevel: l}, nil
}

func run(c *config) error {
	log.Info("running...")
	lastRestart := &restartConfig{LastRestart: &time.Time{}, ThrottleThresh: time.Second}
	restartSystemdServiceThrottledAndLog(c.SystemdService, lastRestart)
	err := onChange(c.ConfigFilePath, func() { restartSystemdServiceThrottledAndLog(c.SystemdService, lastRestart) })
	if err != nil {
		return wrapAndTrace(err)
	}
	log.Info("successful.")
	return nil
}

type restartConfig struct {
	LastRestart    *time.Time
	ThrottleThresh time.Duration
}

func restartSystemdServiceThrottledAndLog(systemdService string, config *restartConfig) {
	defer func() {
		now := time.Now()
		config.LastRestart = &now
	}()

	log.Debugf("Time since last restart %v", time.Since(*config.LastRestart))
	if time.Since(*config.LastRestart) < config.ThrottleThresh {
		log.Info("throttling restart")
		return
	}

	err := restartSystemdService(systemdService)
	if err != nil {
		log.Error(err)
	}
}

func restartSystemdService(systemdService string) error {
	log.Info("restarting vault")
	cmd := exec.Command("systemctl", "restart", systemdService) //nolint // accepting vul risk
	if stdout, err := cmd.Output(); err != nil {
		return wrapAndTrace(err, string(stdout))
	}

	return nil
}

func onChange(configFilePath string, action func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return wrapAndTrace(err)
	}
	defer checks(watcher.Close)

	done := make(chan bool)

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				log.Tracef("EVENT %v\n", event)
				action()
			// watch for errors
			case err := <-watcher.Errors:
				log.Warnf("ERROR %v\n", err)
			}
		}
	}()

	if err := watcher.Add(configFilePath); err != nil {
		return wrapAndTrace(err)
	}

	<-done
	return nil
}

func wrapAndTrace(err error, messages ...string) error {
	message := ""
	for _, m := range messages {
		message += fmt.Sprintf(" %s", m)
	}
	return errors.Wrap(err, makeErrorMessage(message))
}

func makeErrorMessage(message string) string {
	errorCall := 2
	_, fn, line, _ := runtime.Caller(errorCall)
	return fmt.Sprintf("[error] %s:%d %s\n\t", fn, line, message)
}

func checks(fs ...func() error) {
	for i := len(fs) - 1; i >= 0; i-- {
		if err := fs[i](); err != nil {
			log.Error(err)
		}
	}
}
