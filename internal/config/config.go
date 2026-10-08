package config

//go:generate go tool configulator -type Config

type Config struct {
	LogLevel           LogLevel    `name:"log-level" description:"Logging level for the application. One of debug, info, warn, or error" default:"info"`
	Port               int         `name:"port" description:"Port to listen on" default:"8080"`
	MQTT               MQTT        `name:"mqtt" description:"MQTT configuration"`
	Image              Image       `name:"image" description:"Image configuration"`
	CORSAllowedOrigins []string    `name:"cors-allowed-origins" description:"CORS allowed origins" default:"https://*,http://*"`
	PublicFrame        PublicFrame `name:"public-frame" description:"The observatory's own newest sub of the target, from astro-stacker"`
}

// PublicFrame shows the observatory's newest sub of the target being
// imaged, as astro-stacker renders it for the public, instead of the survey
// image. Off without a stacker URL.
type PublicFrame struct {
	StackerURL      string `name:"stacker-url" description:"astro-stacker's base URL, e.g. http://astro-stacker.astro-processing:8080; empty shows only survey images"`
	PublicURL       string `name:"public-url" description:"This service's public base URL, which the page loads the frame from" default:"https://wheresmyscope.mcswain.dev"`
	IntervalSeconds int    `name:"interval-seconds" description:"Seconds between checks for a newer frame" default:"60"`
	TimeoutSeconds  int    `name:"timeout-seconds" description:"Seconds before a fetch from the stacker gives up and the survey image is shown" default:"10"`
}

type Image struct {
	Projection ProjectionType `name:"projection" description:"Projection type" default:"STG"`
	FOV        float64        `name:"fov" description:"Field of view in degrees" default:"3.3"`
	Format     ImageFormat    `name:"format" description:"Image format" default:"png"`
	Width      int            `name:"width" description:"Image width in pixels" default:"900"`
	Height     int            `name:"height" description:"Image height in pixels" default:"600"`
	Stretch    StretchType    `name:"stretch" description:"Stretch type" default:"linear"`
	MinCut     float64        `name:"min-cut" description:"Minimum cut value for image processing" default:"0.5"`
	MaxCut     float64        `name:"max-cut" description:"Maximum cut value for image processing" default:"99.5"`
	HiPS       string         `name:"hips" description:"HIPS name for the image" default:"CDS/P/DSS2/color"`
}

type MQTT struct {
	Broker   string `name:"broker" description:"MQTT broker address"`
	ClientID string `name:"client-id" description:"Client ID for MQTT connection" default:"wheresmyscope"`
	Prefix   string `name:"prefix" description:"Prefix for MQTT topics" default:"wheresmyscope"`
	Username string `name:"username" description:"Username for MQTT connection"`
	Password string `name:"password" description:"Password for MQTT connection"`
}
