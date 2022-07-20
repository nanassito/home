package mqtt

import (
	"flag"

	paho "github.com/eclipse/paho.mqtt.golang"
)

var server = flag.String("mqtt", "tcp://192.168.1.1:1883", "Address of the mqtt server.")

type MqttIface interface {
	Reset()
	PublishString(topic string, message string) error
}

type Mqtt struct {
	clientID string
	client   paho.Client
}

func newClient(clientID string) paho.Client {
	client := paho.NewClient(paho.NewClientOptions().SetClientID(clientID).AddBroker(*server))
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return client
}

func New(clientID string) *Mqtt {
	flag.Parse()
	return &Mqtt{clientID: clientID, client: newClient(clientID)}
}

func (m *Mqtt) Reset() {
	m.client = newClient(m.clientID)
}

func (m *Mqtt) PublishString(topic string, message string) error {
	t := m.client.Publish(topic, 0, false, message)
	<-t.Done()
	return t.Error()
}
