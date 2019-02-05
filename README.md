# mqttlog

A logging backend for "github.com/op/go-logging" to send messages to one or more MQTT topics

## Synopsis

This Go package is designed to be used in conjunction with:

- Paho MQTT Client: [github.com/eclipse/paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang)
- Golang Logging Library: [github.com/op/go-logging](https://github.com/op/go-logging)

The idea here is that you might find logging messages in real time to an MQTT broker to be a useful approach to distributed log management.

For example; you have multiple services or components in your complete application stack (front end web, mobile and backend), and you wish to aggregate your logs. Thus, allowing you to create your own alerting, self healing, log visualization, debugging and performance metrics tooling.

The nature of MQTT topic structures enables you to devise a "best of both worlds" approach to separated AND combined log streams. And utilising MQTT's [single and multi level wildcards](https://www.hivemq.com/blog/mqtt-essentials-part-5-mqtt-topics-best-practices/) can make viewing & processing logs very convenient indeed.

## Installation

```
go get github.com/camstuart/mqttlog
```

## TLDR; Example

```Golang
logger := logging.MustGetLogger("main")
mqttBackend := mqttlog.NewBackend(mqc, "logging/text") # mqc is a github.com/eclipse/paho.mqtt.golang.Client
logging.SetBackend(mqttBackend)
```

## Full Example

Below is a complete example of initialising your application with a connection to an MQTT broker, and setting up multiple logging backends so a single logging call results in the log message being emitted to multiple MQTT topics.

Note: MQTT topic names are arbitrary, below is an example of my personal naming convention

```Golang
package main

import (
	"fmt"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	logging "github.com/op/go-logging"
	"github.com/camstuart/mqttlog"
)

var (
	serviceID = "test-golang-mqtt-logger"
	appID     = "my-big-app"
	mqc       MQTT.Client
	logger    = logging.MustGetLogger("main")
)

func init() {

	logFormatter := logging.MustStringFormatter(`%{color}%{time:15:04:05.000} %{program} %{shortfunc} %{level:.10s} %{id:03x}%{color:reset} %{message}`)
	logging.SetFormatter(logFormatter)
	consoleBackend := logging.NewLogBackend(os.Stdout, "", 0)
	consoleBackend.Color = true
	logging.SetLevel(logging.INFO, "main")
	logging.SetBackend(consoleBackend)

	opts := MQTT.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883")
	opts.SetClientID(serviceID)
	opts.SetUsername("myusername")
	opts.SetPassword("password")

	mqc = MQTT.NewClient(opts)
	if token := mqc.Connect(); token.Wait() && token.Error() != nil {
		logger.Errorf("unable to establish mqtt broker connection, error was: %s", token.Error())
	}

	mqttMinimalTextBackend := mqttlog.NewBackend(mqc, fmt.Sprintf("%s/logging/%s/minimal/text", appID, serviceID), mqttlog.StyleMinimal)
	mqttFormattedTextBackend := mqttlog.NewBackend(mqc, fmt.Sprintf("%s/logging/%s/formatted/text", appID, serviceID), mqttlog.StyleFormatted)
	mqttFormattedJSONBackend := mqttlog.NewBackend(mqc, fmt.Sprintf("%s/logging/%s/formatted/json", appID, serviceID), mqttlog.StyleJSON)
	logging.SetBackend(consoleBackend, mqttMinimalTextBackend, mqttFormattedTextBackend, mqttFormattedJSONBackend)
}

func main() {
	defer mqc.Disconnect(250)
	logger.Info("test INFO log message")
	logger.Debug("test INFO log message")
	logger.Warning("test WARN log message")
}
```