package config

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

type ImageFormat string

const (
	ImageFormatPNG  ImageFormat = "png"
	ImageFormatJPEG ImageFormat = "jpeg"
	ImageFormatFITS ImageFormat = "fits"
)

type StretchType string

const (
	StretchTypePower  StretchType = "power"
	StretchTypeLinear StretchType = "linear"
	StretchTypeSqrt   StretchType = "sqrt"
	StretchTypeLog    StretchType = "log"
	StretchTypeAsinh  StretchType = "asinh"
)
