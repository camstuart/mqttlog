package mqttlog

import (
	"encoding/json"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	logging "github.com/op/go-logging"
)

// MQTTBackend sends log messages to the nominated MQTT broker
type MQTTBackend struct {
	broker MQTT.Client
	topic  string
	style  string
}

// NewBackend creates a new MQTTBackend.
func NewBackend(brokerClient MQTT.Client, topic, style string) *MQTTBackend {
	mb := &MQTTBackend{
		broker: brokerClient,
		topic:  topic,
		style:  style,
	}
	return mb
}

// Log implements the logging.Backend interface.
func (b *MQTTBackend) Log(level logging.Level, calldepth int, rec *logging.Record) error {
	var message string

	switch b.style {
	case StyleJSON:
		msg := LogMessage{
			ID:      rec.ID,
			Time:    rec.Time,
			Message: rec.Message(),
			Level:   rec.Level.String(),
			Module:  rec.Module,
		}
		mb, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		message = string(mb)
	case StyleMinimal:
		message = rec.Message()
	default:
		message = rec.Formatted(calldepth + 1)
	}

	token := b.broker.Publish(b.topic, byte(0), false, []byte(message))
	token.WaitTimeout(time.Second * 5)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}
	return nil
}

const (
	// StyleFormatted uses the existing logging format untouched, or uses the default if none was specified
	StyleFormatted = "FORMATTED"
	// StyleMinimal outputs only the message
	StyleMinimal = "MINIMAL"
	// StyleJSON outputs the message as a JSON string for ease of parsing
	StyleJSON = "JSON"
)

// LogMessage is a struct to represent a full log entry, used to marshall into JSON
// When StyleJSON is used
type LogMessage struct {
	ID        uint64    `json:"id,omitempty"`
	PID       int       `json:"pid,omitempty"`
	Time      time.Time `json:"time,omitempty"`
	Level     string    `json:"level,omitempty"`
	Module    string    `json:"module,omitempty"`
	Program   string    `json:"program,omitempty"`
	Message   string    `json:"message,omitempty"`
	LongFile  string    `json:"longFileName,omitempty"`
	ShortFile string    `json:"shortFileName,omitempty"`
	CallPath  string    `json:"callPath,omitempty"`
}
