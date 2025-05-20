package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/USA-RedDragon/wheresmyscope/internal/config"
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
	client    *autopaho.ConnectionManager
	config    *config.Config
	state     ScopeState
	stateLock sync.Mutex
}

func NewMQTT(ctx context.Context, config *config.Config) (*MQTT, error) {
	u, err := url.Parse(config.MQTT.Broker)
	if err != nil {
		return nil, err
	}

	mqtt := &MQTT{
		config: config,
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

	mqtt.client = c
	return mqtt, nil
}

func (m *MQTT) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

func (m *MQTT) GetState() ScopeState {
	return m.state
}

func (m *MQTT) updateState(topic, payload string) {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()

	switch topic {
	case m.config.MQTT.Prefix + "/name":
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
		_, err = m.client.Publish(context.Background(), &paho.Publish{
			Topic:   m.config.MQTT.Prefix + "/ra_decimal_degrees",
			QoS:     1,
			Retain:  true,
			Payload: []byte(fmt.Sprintf("%f", m.state.RightAscension)),
		})
		if err != nil {
			slog.Error("failed to publish RA", "error", err)
		}
	case m.config.MQTT.Prefix + "/dec_decimal":
		dec, err := strconv.ParseFloat(payload, 64)
		if err == nil {
			m.state.Declination = dec
		} else {
			slog.Error("failed to parse DEC", "error", err)
		}
		_, err = m.client.Publish(context.Background(), &paho.Publish{
			Topic:   m.config.MQTT.Prefix + "/dec_decimal_degrees",
			QoS:     1,
			Retain:  true,
			Payload: []byte(fmt.Sprintf("%f", m.state.Declination)),
		})
		if err != nil {
			slog.Error("failed to publish DEC", "error", err)
		}
	case m.config.MQTT.Prefix + "/available":
		m.state.Live = payload == "true"
	default:
		return
	}

	queryParams := url.Values{}
	queryParams.Set("projection", string(m.config.Image.Projection))
	queryParams.Set("hips", m.config.Image.HiPS)
	queryParams.Set("fov", fmt.Sprintf("%f", m.config.Image.FOV))
	queryParams.Set("ra", fmt.Sprintf("%f", m.state.RightAscension))
	queryParams.Set("dec", fmt.Sprintf("%f", m.state.Declination))
	queryParams.Set("format", string(m.config.Image.Format))
	queryParams.Set("width", fmt.Sprintf("%d", m.config.Image.Width))
	queryParams.Set("height", fmt.Sprintf("%d", m.config.Image.Height))
	queryParams.Set("stretch", string(m.config.Image.Stretch))
	queryParams.Set("rotation_angle", fmt.Sprintf("%f", m.state.Rotation))
	queryParams.Set("min_cut", fmt.Sprintf("%f%%", m.config.Image.MinCut))
	queryParams.Set("max_cut", fmt.Sprintf("%f%%", m.config.Image.MaxCut))

	url := "https://alaskybis.u-strasbg.fr/hips-image-services/hips2fits"
	queryString := queryParams.Encode()
	if queryString != "" {
		url += "?" + queryString
	}

	m.state.ImageURL = url

	_, err := m.client.Publish(context.Background(), &paho.Publish{
		Topic:   m.config.MQTT.Prefix + "/image_url",
		QoS:     1,
		Retain:  true,
		Payload: []byte(m.state.ImageURL),
	})
	if err != nil {
		slog.Error("failed to publish image URL", "error", err)
	}

}
