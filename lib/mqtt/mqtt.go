package mqtt

import (
	"flag"

	paho "github.com/eclipse/paho.mqtt.golang"
)

var server = flag.String("mqtt", "tcp://192.168.1.1:1883", "Address of the mqtt server.")

type MqttIface interface {
	PublishString(topic string, message string) error
}

type Mqtt struct {
	client paho.Client
}

func New(ClientID string) *Mqtt {
	flag.Parse()
	client := paho.NewClient(paho.NewClientOptions().SetClientID(ClientID).AddBroker(*server))
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return &Mqtt{client: client}
}

func (m *Mqtt) PublishString(topic string, message string) error {
	t := m.client.Publish(topic, 0, false, message)
	<-t.Done()
	return t.Error()
}
