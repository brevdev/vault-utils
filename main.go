package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/fsnotify/fsnotify"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type config struct {
	SystemdService string
	ConfigFilePath string
	LogLevel       LogLevel
}

func (c config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.SystemdService, validation.Required),
		validation.Field(&c.ConfigFilePath, validation.Required, is.RequestURI),
		validation.Field(&c.LogLevel),
	)
}

// LogLevel error, warn, info, debug, etc.
type LogLevel string

// Validate checks if log is proper.
func (l LogLevel) Validate() error {
	_, err := log.ParseLevel(string(l))
	return err
}

func main() {
	c, err := getConfig()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	level, _ := log.ParseLevel(string(c.LogLevel)) // validation already has taken place
	log.SetLevel(level)

	if err = run(c); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func getConfig() (*config, error) {
	systemdService := flag.String("service", "", "The systemd service to restart")
	configFilePath := flag.String("configPath", "", "Path to config file to watch")
	logLevel := flag.String("logLevel", "info", "Ex: trace, debug, info, warn, error")
	flag.Parse()

	c := config{SystemdService: *systemdService, ConfigFilePath: *configFilePath, LogLevel: LogLevel(*logLevel)}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &c, nil
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
	if stdout, err := cmd.CombinedOutput(); err != nil {
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
