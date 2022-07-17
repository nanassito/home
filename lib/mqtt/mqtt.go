package mqtt

import (
	"flag"
	"net/url"

	paho "github.com/eclipse/paho.mqtt.golang"
)

var server = flag.String("mqtt", "192.168.1.1:1883", "Address of the mqtt server.")

type MqttIface interface {
	PublishString(topic string, message string) error
}

type Mqtt struct {
	client paho.Client
}

func New(ClientID string) *Mqtt {
	flag.Parse()
	client := paho.NewClient(&paho.ClientOptions{
		Servers: []*url.URL{
			{
				User: &url.Userinfo{},
				Host: *server,
			},
		},
		ClientID: ClientID,
	})
	return &Mqtt{client: client}
}

func (m *Mqtt) PublishString(topic string, message string) error {
	m.client.Connect() // TODO: Mqtt failure: not Connected  ????
	t := m.client.Publish(topic, 0, false, message)
	<-t.Done()
	return t.Error()
}
