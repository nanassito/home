package mqtt

import (
	"flag"
	"fmt"
	"log"
	"os"

	paho "github.com/eclipse/paho.mqtt.golang"
)

var (
	server   = flag.String("mqtt", "tcp://192.168.1.1:1883", "Address of the mqtt server.")
	logger   = log.New(os.Stderr, "", log.Lshortfile)
	hostname string
)

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		panic(fmt.Sprintf("Can't figure out the hostname: %v", err))
	}
}

type MqttIface interface {
	Reset()
	PublishString(topic string, message string) error
	Subscribe(topic string, callback func(topic string, payload []byte)) error
}

type Mqtt struct {
	clientID string
	client   paho.Client
}

func newClient(clientID string) paho.Client {
	opts := paho.NewClientOptions()
	opts.SetClientID(clientID + "-" + hostname)
	opts.AddBroker(*server)
	opts.OnConnectionLost = func(client paho.Client, err error) {
		logger.Printf("Lost mqtt connection: %v", err)
	}
	client := paho.NewClient(opts)
	logger.Printf("Info| Connecting to Mqtt broker at %s\n", *server)
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

func (m *Mqtt) Subscribe(topic string, callback func(topic string, payload []byte)) error {
	t := m.client.Subscribe(topic, 1, func(client paho.Client, message paho.Message) {
		callback(message.Topic(), message.Payload())
	})
	<-t.Done()
	logger.Printf("Info| Subscribing to mqtt://%s\n", topic)
	err := t.Error()
	if err != nil {
		logger.Printf("Fail| Failed to subscribe to %s\n", topic)
	}
	return err
}
