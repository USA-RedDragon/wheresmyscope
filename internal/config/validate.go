package config

import "errors"

var (
	ErrInvalidLogLevel    = errors.New("invalid log level provided")
	ErrInvalidPort        = errors.New("port must be between 1 and 65535")
	ErrNoMQTTBroker       = errors.New("no MQTT broker provided")
	ErrInvalidProjection  = errors.New("invalid projection type provided")
	ErrFOVTooSmall        = errors.New("FOV must be greater than 0")
	ErrInvalidImageFormat = errors.New("invalid image format provided")
	ErrInvalidWidth       = errors.New("image width must be greater than 0")
	ErrInvalidHeight      = errors.New("image height must be greater than 0")
	ErrInvalidStretch     = errors.New("invalid stretch type provided")
	ErrMinCutTooSmall     = errors.New("min cut must be greater than 0")
	ErrMaxCutTooSmall     = errors.New("max cut must be greater than 0")
	ErrMaxCutTooLarge     = errors.New("max cut must be less than 100")
	ErrMinCutTooLarge     = errors.New("min cut must be less than 100")
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

	if c.Image.FOV <= 0 {
		return ErrFOVTooSmall
	}

	if c.Image.Width <= 0 {
		return ErrInvalidWidth
	}

	if c.Image.Height <= 0 {
		return ErrInvalidHeight
	}

	if c.Image.MinCut <= 0 {
		return ErrMinCutTooSmall
	}

	if c.Image.MaxCut <= 0 {
		return ErrMaxCutTooSmall
	}

	if c.Image.MaxCut >= 100 {
		return ErrMaxCutTooLarge
	}

	if c.Image.MinCut >= 100 {
		return ErrMinCutTooLarge
	}

	if c.Image.Projection != ProjectionZenithalPerspective &&
		c.Image.Projection != ProjectionSlantZenithalPerspective &&
		c.Image.Projection != ProjectionTangential &&
		c.Image.Projection != ProjectionStereographic &&
		c.Image.Projection != ProjectionOrthographic &&
		c.Image.Projection != ProjectionAzimuthalEquidistant &&
		c.Image.Projection != ProjectionZenithalEqualArea &&
		c.Image.Projection != ProjectionAiry &&
		c.Image.Projection != ProjectionCylindricalPerspective &&
		c.Image.Projection != ProjectionCylindricalEqualArea &&
		c.Image.Projection != ProjectionPlateCarree &&
		c.Image.Projection != ProjectionMercator &&
		c.Image.Projection != ProjectionSansonFlamsteed &&
		c.Image.Projection != ProjectionParabolic &&
		c.Image.Projection != ProjectionMollweide &&
		c.Image.Projection != ProjectionHammerAitoff &&
		c.Image.Projection != ProjectionTangentialSphericalCube &&
		c.Image.Projection != ProjectionQuadrilateralizedSphericalCube &&
		c.Image.Projection != ProjectionHEALPix &&
		c.Image.Projection != ProjectionHealPixPolarButterfly {
		return ErrInvalidProjection
	}

	if c.Image.Format != ImageFormatPNG &&
		c.Image.Format != ImageFormatJPEG &&
		c.Image.Format != ImageFormatFITS {
		return ErrInvalidImageFormat
	}

	if c.Image.Stretch != StretchTypePower &&
		c.Image.Stretch != StretchTypeLinear &&
		c.Image.Stretch != StretchTypeSqrt &&
		c.Image.Stretch != StretchTypeLog &&
		c.Image.Stretch != StretchTypeAsinh {
		return ErrInvalidStretch
	}

	return nil
}
