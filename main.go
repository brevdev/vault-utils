package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"crypto/md5" //nolint // using md5 for simple hash

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Config is the parameters that configures vault-utils.
type Config struct {
	SystemdService string
	ConfigFilePath string
	PollTime       Duration // in seconds
	LogLevel       LogLevel
}

// Validate is the config validator.
func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.SystemdService, validation.Required),
		validation.Field(&c.ConfigFilePath, validation.Required, is.RequestURI),
		validation.Field(&c.PollTime),
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

// Duration 1s, 1hr etc.
type Duration string

// Validate checks if log is proper.
func (d Duration) Validate() error {
	_, err := time.ParseDuration(string(d))
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

func getConfig() (*Config, error) {
	systemdService := flag.String("service", "", "The systemd service to restart")
	configFilePath := flag.String("configPath", "", "Path to config file to watch")
	logLevel := flag.String("logLevel", "info", "Ex: trace, debug, info, warn, error")
	pollTime := flag.String("pollTime", "2s", "int in seconds")

	flag.Parse()

	c := Config{SystemdService: *systemdService, ConfigFilePath: *configFilePath, LogLevel: LogLevel(*logLevel), PollTime: Duration(*pollTime)}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &c, nil
}

func run(c *Config) error {
	log.Info("running...")
	lastRestart := &restartConfig{LastRestart: &time.Time{}, ThrottleThresh: time.Second}
	restartSystemdServiceThrottledAndLog(c.SystemdService, lastRestart)

	pollTime, _ := time.ParseDuration(string(c.PollTime)) // validation already happened

	err := onChange(c.ConfigFilePath, func() { restartSystemdServiceThrottledAndLog(c.SystemdService, lastRestart) }, pollTime)
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
	log.Infof("restarting %s", systemdService)
	cmd := exec.Command("systemctl", "restart", systemdService) //nolint // accepting vul risk
	if stdout, err := cmd.CombinedOutput(); err != nil {
		return wrapAndTrace(err, string(stdout))
	}

	return nil
}

func onChange(configFilePath string, action func(), pollTime time.Duration) error {
	var hash string
	for {
		newHash, err := md5sum(configFilePath)
		if err != nil {
			return wrapAndTrace(err)
		}
		if hash != newHash {
			action()
		}
		hash = newHash
		time.Sleep(pollTime)
	}
}

func md5sum(filePath string) (string, error) {
	file, err := os.Open(filePath) //nolint // accepting risk of opening var path
	if err != nil {
		return "", err
	}
	defer checks(file.Close)

	hash := md5.New() //nolint // not using md5 for anything cryptographically important
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
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
