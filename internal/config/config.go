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

type ProjectionType string

const (
	ProjectionZenithalPerspective            ProjectionType = "AZP"
	ProjectionSlantZenithalPerspective       ProjectionType = "SZP"
	ProjectionTangential                     ProjectionType = "TAN"
	ProjectionStereographic                  ProjectionType = "STG"
	ProjectionOrthographic                   ProjectionType = "SIN"
	ProjectionAzimuthalEquidistant           ProjectionType = "ARC"
	ProjectionZenithalEqualArea              ProjectionType = "ZEA"
	ProjectionAiry                           ProjectionType = "AIR"
	ProjectionCylindricalPerspective         ProjectionType = "CYP"
	ProjectionCylindricalEqualArea           ProjectionType = "CEA"
	ProjectionPlateCarree                    ProjectionType = "CAR"
	ProjectionMercator                       ProjectionType = "MER"
	ProjectionSansonFlamsteed                ProjectionType = "SFL"
	ProjectionParabolic                      ProjectionType = "PAR"
	ProjectionMollweide                      ProjectionType = "MOL"
	ProjectionHammerAitoff                   ProjectionType = "AIT"
	ProjectionTangentialSphericalCube        ProjectionType = "TSC"
	ProjectionQuadrilateralizedSphericalCube ProjectionType = "QSC"
	ProjectionHEALPix                        ProjectionType = "HPX"
	ProjectionHealPixPolarButterfly          ProjectionType = "XPH"
)

type Config struct {
	LogLevel   LogLevel       `name:"log-level" description:"Logging level for the application. One of debug, info, warn, or error" default:"info"`
	Port       int            `name:"port" description:"Port to listen on" default:"8080"`
	FOV        float64        `name:"fov" description:"Field of view in degrees" default:"3.3"`
	Projection ProjectionType `name:"projection" description:"Projection type" default:"STG"`
	MQTT       MQTT           `name:"mqtt" description:"MQTT configuration"`
}

type MQTT struct {
	Broker   string `name:"broker" description:"MQTT broker address"`
	ClientID string `name:"client-id" description:"Client ID for MQTT connection" default:"wheresmyscope"`
	Prefix   string `name:"prefix" description:"Prefix for MQTT topics" default:"wheresmyscope"`
	Username string `name:"username" description:"Username for MQTT connection"`
	Password string `name:"password" description:"Password for MQTT connection"`
}

var (
	ErrInvalidLogLevel   = errors.New("invalid log level provided")
	ErrInvalidPort       = errors.New("port must be between 1 and 65535")
	ErrNoMQTTBroker      = errors.New("no MQTT broker provided")
	ErrInvalidProjection = errors.New("invalid projection type provided")
	ErrFOVTooSmall       = errors.New("FOV must be greater than 0")
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

	if c.FOV <= 0 {
		return ErrFOVTooSmall
	}

	if c.Projection != ProjectionZenithalPerspective &&
		c.Projection != ProjectionSlantZenithalPerspective &&
		c.Projection != ProjectionTangential &&
		c.Projection != ProjectionStereographic &&
		c.Projection != ProjectionOrthographic &&
		c.Projection != ProjectionAzimuthalEquidistant &&
		c.Projection != ProjectionZenithalEqualArea &&
		c.Projection != ProjectionAiry &&
		c.Projection != ProjectionCylindricalPerspective &&
		c.Projection != ProjectionCylindricalEqualArea &&
		c.Projection != ProjectionPlateCarree &&
		c.Projection != ProjectionMercator &&
		c.Projection != ProjectionSansonFlamsteed &&
		c.Projection != ProjectionParabolic &&
		c.Projection != ProjectionMollweide &&
		c.Projection != ProjectionHammerAitoff &&
		c.Projection != ProjectionTangentialSphericalCube &&
		c.Projection != ProjectionQuadrilateralizedSphericalCube &&
		c.Projection != ProjectionHEALPix &&
		c.Projection != ProjectionHealPixPolarButterfly {
		return ErrInvalidProjection
	}

	return nil
}
