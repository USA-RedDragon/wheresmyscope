package config

import (
	"errors"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type Config struct {
	LogLevel LogLevel `name:"log-level" description:"Logging level for the application. One of debug, info, warn, or error" default:"info"`
	Port     int      `name:"port" description:"Port to listen on" default:"8080"`
	MQTT     MQTT     `name:"mqtt" description:"MQTT configuration"`
}

type MQTT struct {
	Broker   string `name:"broker" description:"MQTT broker address"`
	ClientID string `name:"client-id" description:"Client ID for MQTT connection" default:"wheresmyscope"`
	Prefix   string `name:"prefix" description:"Prefix for MQTT topics" default:"wheresmyscope"`
	Username string `name:"username" description:"Username for MQTT connection"`
	Password string `name:"password" description:"Password for MQTT connection"`
}

var (
	ErrInvalidLogLevel = errors.New("invalid log level provided")
	ErrInvalidPort     = errors.New("port must be between 1 and 65535")
	ErrNoMQTTBroker    = errors.New("no MQTT broker provided")
)

func (c Config) Validate() error {
	if c.LogLevel != LogLevelDebug &&
		c.LogLevel != LogLevelInfo &&
		c.LogLevel != LogLevelWarn &&
		c.LogLevel != LogLevelError {
		return ErrInvalidLogLevel
	}

	if c.Port < 1 || c.Port > 65535 {
		return ErrInvalidPort
	}

	if c.MQTT.Broker == "" {
		return ErrNoMQTTBroker
	}

	return nil
}
