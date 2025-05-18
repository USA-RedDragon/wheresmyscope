package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/USA-RedDragon/wheresmyscope/internal/config"
	"github.com/USA-RedDragon/wheresmyscope/internal/utils"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type ScopeState struct {
	Target         string    `json:"target"`
	Start          time.Time `json:"start"`
	RightAscension float64   `json:"ra"`
	Declination    float64   `json:"dec"`
	Available      bool      `json:"available"`
	ImageURL       string    `json:"image_url"`
}

type MQTT struct {
	client    *autopaho.ConnectionManager
	config    *config.MQTT
	state     ScopeState
	stateLock sync.Mutex
}

func NewMQTT(ctx context.Context, config *config.MQTT) (*MQTT, error) {
	u, err := url.Parse(config.Broker)
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
		ConnectUsername:       config.Username,
		ConnectPassword:       []byte(config.Password),
		ClientConfig: paho.ClientConfig{
			ClientID: config.ClientID,
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
				Topic: config.Prefix + "/#",
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
	case m.config.Prefix + "/name":
		m.state.Target = payload
	case m.config.Prefix + "/start":
		start, err := time.Parse(time.RFC3339, payload)
		if err == nil {
			m.state.Start = start
		}
	case m.config.Prefix + "/ra":
		ra, err := utils.RaToDegrees(payload)
		if err == nil {
			m.state.RightAscension = ra
		} else {
			slog.Error("failed to parse RA", "error", err)
		}
		_, err = m.client.Publish(context.Background(), &paho.Publish{
			Topic:   m.config.Prefix + "/ra_decimal_degrees",
			QoS:     1,
			Retain:  true,
			Payload: []byte(fmt.Sprintf("%f", m.state.RightAscension)),
		})
	case m.config.Prefix + "/dec":
		dec, err := utils.DecToDegrees(payload)
		if err == nil {
			m.state.Declination = dec
		} else {
			slog.Error("failed to parse DEC", "error", err)
		}
		_, err = m.client.Publish(context.Background(), &paho.Publish{
			Topic:   m.config.Prefix + "/dec_decimal_degrees",
			QoS:     1,
			Retain:  true,
			Payload: []byte(fmt.Sprintf("%f", m.state.Declination)),
		})
		if err != nil {
			slog.Error("failed to publish DEC", "error", err)
		}
	case m.config.Prefix + "/available":
		m.state.Available = payload == "true"
	default:
		return
	}

	m.state.ImageURL = fmt.Sprintf(
		"https://alaskybis.u-strasbg.fr/hips-image-services/hips2fits?projection=STG&hips=CDS%%2FP%%2FDSS2%%2Fcolor&width=900&height=600&fov=3.3&ra=%f&dec=%f&format=jpg",
		m.state.RightAscension, m.state.Declination,
	)

	_, err := m.client.Publish(context.Background(), &paho.Publish{
		Topic:   m.config.Prefix + "/image_url",
		QoS:     1,
		Retain:  true,
		Payload: []byte(m.state.ImageURL),
	})
	if err != nil {
		slog.Error("failed to publish image URL", "error", err)
	}

}
