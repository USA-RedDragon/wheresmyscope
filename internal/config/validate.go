package config

import (
	"errors"
	"fmt"
	"net/netip"
)

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
	ErrPublicFrameTiming  = errors.New("public frame interval and timeout must be greater than 0")
	ErrNoPublicURL        = errors.New("public frames need this service's public URL")
	ErrInvalidProxy       = errors.New("trusted proxy must be an IP or CIDR")
)

// TrustedProxyPrefixes parses TrustedProxies. A bare IP becomes a
// single-address prefix.
func (c Config) TrustedProxyPrefixes() ([]netip.Prefix, error) {
	out := make([]netip.Prefix, 0, len(c.TrustedProxies))
	for _, p := range c.TrustedProxies {
		if prefix, err := netip.ParsePrefix(p); err == nil {
			out = append(out, prefix.Masked())
			continue
		}
		addr, err := netip.ParseAddr(p)
		if err != nil {
			return nil, fmt.Errorf("%w: %q", ErrInvalidProxy, p)
		}
		out = append(out, netip.PrefixFrom(addr.Unmap(), addr.Unmap().BitLen()))
	}
	return out, nil
}

func (c Config) Validate() error {
	if !c.LogLevel.valid() {
		return ErrInvalidLogLevel
	}

	if _, err := c.TrustedProxyPrefixes(); err != nil {
		return err
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

	if !c.Image.Projection.valid() {
		return ErrInvalidProjection
	}

	if !c.Image.Format.valid() {
		return ErrInvalidImageFormat
	}

	if !c.Image.Stretch.valid() {
		return ErrInvalidStretch
	}

	if c.PublicFrame.StackerURL != "" {
		if c.PublicFrame.IntervalSeconds <= 0 || c.PublicFrame.TimeoutSeconds <= 0 {
			return ErrPublicFrameTiming
		}
		if c.PublicFrame.PublicURL == "" {
			return ErrNoPublicURL
		}
	}

	return nil
}

func (l LogLevel) valid() bool {
	switch l {
	case
		LogLevelDebug,
		LogLevelInfo,
		LogLevelWarn,
		LogLevelError:
		return true
	}
	return false
}

func (p ProjectionType) valid() bool {
	switch p {
	case
		ProjectionZenithalPerspective,
		ProjectionSlantZenithalPerspective,
		ProjectionTangential,
		ProjectionStereographic,
		ProjectionOrthographic,
		ProjectionAzimuthalEquidistant,
		ProjectionZenithalEqualArea,
		ProjectionAiry,
		ProjectionCylindricalPerspective,
		ProjectionCylindricalEqualArea,
		ProjectionPlateCarree,
		ProjectionMercator,
		ProjectionSansonFlamsteed,
		ProjectionParabolic,
		ProjectionMollweide,
		ProjectionHammerAitoff,
		ProjectionTangentialSphericalCube,
		ProjectionQuadrilateralizedSphericalCube,
		ProjectionHEALPix,
		ProjectionHealPixPolarButterfly:
		return true
	}
	return false
}

func (f ImageFormat) valid() bool {
	switch f {
	case
		ImageFormatPNG,
		ImageFormatJPEG,
		ImageFormatFITS:
		return true
	}
	return false
}

func (s StretchType) valid() bool {
	switch s {
	case
		StretchTypePower,
		StretchTypeLinear,
		StretchTypeSqrt,
		StretchTypeLog,
		StretchTypeAsinh:
		return true
	}
	return false
}
