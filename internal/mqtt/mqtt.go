package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/USA-RedDragon/wheresmyscope/internal/config"
	"github.com/USA-RedDragon/wheresmyscope/internal/publicframe"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
)

type ScopeState struct {
	Target         string    `json:"target"`
	Start          time.Time `json:"start"`
	Rotation       float64   `json:"rotation"`
	RightAscension float64   `json:"ra"`
	Declination    float64   `json:"dec"`
	Live           bool      `json:"live"`
	ImageURL       string    `json:"image_url"`
}

type MQTT struct {
	client     *autopaho.ConnectionManager
	clientLock sync.RWMutex
	config     *config.Config
	state      ScopeState
	stateLock  sync.Mutex

	// surveyURL is the hips2fits cutout of where the scope points; the
	// page shows it unless frames has the observatory's own sub of the
	// target being imaged.
	surveyURL string
	frames    *publicframe.Fetcher
	publicURL string
	// kick asks the frame poller to fetch now: the target or the
	// observatory's availability changed.
	kick chan struct{}
	// published is the image URL last published over MQTT.
	published string
}

func NewMQTT(ctx context.Context, config *config.Config) (*MQTT, error) {
	u, err := url.Parse(config.MQTT.Broker)
	if err != nil {
		return nil, err
	}

	mqtt := &MQTT{
		config: config,
		kick:   make(chan struct{}, 1),
	}
	if config.PublicFrame.StackerURL != "" {
		mqtt.frames = publicframe.New(config.PublicFrame.StackerURL, time.Duration(config.PublicFrame.TimeoutSeconds)*time.Second)
		mqtt.publicURL = strings.TrimSuffix(config.PublicFrame.PublicURL, "/")
	}

	pahoConfig := autopaho.ClientConfig{
		ServerUrls:            []*url.URL{u},
		KeepAlive:             30,
		SessionExpiryInterval: 0xFFFFFFFE, // Never expire
		ConnectUsername:       config.MQTT.Username,
		ConnectPassword:       []byte(config.MQTT.Password),
		ClientConfig: paho.ClientConfig{
			ClientID: fmt.Sprintf("%s_%s", config.MQTT.ClientID, uuid.New().String()),
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					mqtt.updateState(pr.Packet.Topic, string(pr.Packet.Payload))
					return true, nil
				}},
		},
	}

	c, err := autopaho.NewConnection(ctx, pahoConfig)
	if err != nil {
		return nil, err
	}

	if err = c.AwaitConnection(ctx); err != nil {
		return nil, err
	}
	mqtt.clientLock.Lock()
	mqtt.client = c
	mqtt.clientLock.Unlock()

	_, err = c.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic: config.MQTT.Prefix + "/#",
				QoS:   1,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return mqtt, nil
}

func (m *MQTT) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

// GetState is the scope's state, with the image the page should show: the
// observatory's newest sub of the target, or the survey cutout when the
// stacker has none.
func (m *MQTT) GetState() ScopeState {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	s := m.state
	s.ImageURL = m.imageURL()
	return s
}

// Frames serves the held public frame; nil when they are off.
func (m *MQTT) Frames() http.Handler {
	if m.frames == nil {
		return nil
	}
	return m.frames
}

// imageURL is the image to show; the caller holds stateLock.
func (m *MQTT) imageURL() string {
	if m.frames != nil {
		if f, ok := m.frames.Current(m.state.Target); ok {
			return m.publicURL + "/image.jpg?v=" + url.QueryEscape(f.Version())
		}
	}
	return m.surveyURL
}

// RunPublicFrames keeps the public frame of the scope's target current, polling every interval and whenever the target or availability
// changes, until ctx ends. Nothing is fetched per page request.
func (m *MQTT) RunPublicFrames(ctx context.Context, interval time.Duration) {
	if m.frames == nil {
		return
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		m.stateLock.Lock()
		target := m.state.Target
		m.stateLock.Unlock()
		if target != "" {
			if err := m.frames.Fetch(ctx, target); err != nil && ctx.Err() == nil {
				slog.Warn("Could not fetch the public frame", "target", target, "error", err)
			}
		}
		m.publishImageURL(false)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		case <-m.kick:
		}
	}
}

// publishImageURL publishes the image URL to show, if it changed or always.
func (m *MQTT) publishImageURL(always bool) {
	m.stateLock.Lock()
	u := m.imageURL()
	if !always && u == m.published {
		m.stateLock.Unlock()
		return
	}
	m.published = u
	m.stateLock.Unlock()
	m.publish("/image_url", u)
}

// publish sends a retained message under the prefix. Retained messages
// arrive as soon as the subscription is made, so this can run before the
// client is stored; those publishes are skipped (the next update sends
// them).
func (m *MQTT) publish(topic, payload string) {
	m.clientLock.RLock()
	c := m.client
	m.clientLock.RUnlock()
	if c == nil {
		return
	}
	if _, err := c.Publish(context.Background(), &paho.Publish{
		Topic:   m.config.MQTT.Prefix + topic,
		QoS:     1,
		Retain:  true,
		Payload: []byte(payload),
	}); err != nil {
		slog.Error("failed to publish", "topic", topic, "error", err)
	}
}

func (m *MQTT) updateState(topic, payload string) {
	if !m.applyState(topic, payload) {
		return
	}
	m.publishImageURL(true)
}

// applyState records one topic's update and works out the survey URL; it
// reports whether the topic was one of the state's.
func (m *MQTT) applyState(topic, payload string) bool {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()

	switch topic {
	case m.config.MQTT.Prefix + "/name":
		if m.state.Target != payload {
			m.poke()
		}
		m.state.Target = payload
	case m.config.MQTT.Prefix + "/start":
		start, err := time.Parse(time.RFC3339, payload)
		if err == nil {
			m.state.Start = start
		} else {
			slog.Error("failed to parse start time", "error", err)
		}
	case m.config.MQTT.Prefix + "/rotation":
		rotation, err := strconv.ParseFloat(payload, 64)
		if err == nil {
			m.state.Rotation = rotation
		} else {
			slog.Error("failed to parse rotation", "error", err)
		}
	case m.config.MQTT.Prefix + "/ra_decimal":
		val, err := strconv.ParseFloat(payload, 64)
		if err == nil {
			m.state.RightAscension = val * 15 // 1 hour is 15 degrees
		} else {
			slog.Error("failed to parse RA", "error", err)
		}
		m.publish("/ra_decimal_degrees", fmt.Sprintf("%f", m.state.RightAscension))
	case m.config.MQTT.Prefix + "/dec_decimal":
		dec, err := strconv.ParseFloat(payload, 64)
		if err == nil {
			m.state.Declination = dec
		} else {
			slog.Error("failed to parse DEC", "error", err)
		}
		m.publish("/dec_decimal_degrees", fmt.Sprintf("%f", m.state.Declination))
	case m.config.MQTT.Prefix + "/available":
		if live := payload == "true"; live != m.state.Live {
			m.poke()
		}
		m.state.Live = payload == "true"
	default:
		return false
	}

	queryParams := url.Values{}
	queryParams.Set("projection", string(m.config.Image.Projection))
	queryParams.Set("hips", m.config.Image.HiPS)
	queryParams.Set("fov", fmt.Sprintf("%.1f", m.config.Image.FOV))
	queryParams.Set("ra", fmt.Sprintf("%.3f", m.state.RightAscension))
	queryParams.Set("dec", fmt.Sprintf("%.3f", m.state.Declination))
	if m.config.Image.Format != config.ImageFormatFITS {
		queryParams.Set("format", string(m.config.Image.Format))
	}
	queryParams.Set("width", fmt.Sprintf("%d", m.config.Image.Width))
	queryParams.Set("height", fmt.Sprintf("%d", m.config.Image.Height))
	if m.config.Image.Stretch != config.StretchTypeLinear {
		queryParams.Set("stretch", string(m.config.Image.Stretch))
	}
	if m.state.Rotation != 0 {
		queryParams.Set("rotation_angle", fmt.Sprintf("%.2f", m.state.Rotation))
	}
	if m.config.Image.MinCut != 0.5 {
		queryParams.Set("min_cut", fmt.Sprintf("%.1f%%", m.config.Image.MinCut))
	}
	if m.config.Image.MaxCut != 99.5 {
		queryParams.Set("max_cut", fmt.Sprintf("%.1f%%", m.config.Image.MaxCut))
	}

	url := "https://alaskybis.u-strasbg.fr/hips-image-services/hips2fits"
	queryString := queryParams.Encode()
	if queryString != "" {
		url += "?" + queryString
	}

	m.surveyURL = url
	return true
}

// poke wakes the frame poller without blocking.
func (m *MQTT) poke() {
	select {
	case m.kick <- struct{}{}:
	default:
	}
}
