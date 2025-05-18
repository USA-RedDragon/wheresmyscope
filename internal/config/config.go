package config

type Config struct {
	LogLevel LogLevel `name:"log-level" description:"Logging level for the application. One of debug, info, warn, or error" default:"info"`
	Port     int      `name:"port" description:"Port to listen on" default:"8080"`
	MQTT     MQTT     `name:"mqtt" description:"MQTT configuration"`
	Image    Image    `name:"image" description:"Image configuration"`
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
