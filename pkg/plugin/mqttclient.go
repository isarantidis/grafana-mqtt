package plugin

import (
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type MqttClient struct {
	internalClient mqtt.Client
	qos            uint8
	perTopicMutex  *sync.Map
}

type SubscriptionInfo struct {
	mut       *sync.Mutex
	callbacks map[SubscriptionModel]func(TimedMessage)
	counter   uint8
}

func NewMQTTManager(qos uint8, client mqtt.Client) *MqttClient {
	return &MqttClient{
		perTopicMutex:  &sync.Map{},
		internalClient: client,
		qos:            qos,
	}
}

func (client *MqttClient) Connect() error {
	options := client.internalClient.OptionsReader()
	if !client.internalClient.IsConnected() {
		connectionToken := client.internalClient.Connect()
		if connectionToken.Wait() && connectionToken.Error() != nil {
			return fmt.Errorf("could not connect. message: %v", connectionToken.Error())
		}
	} else {
		log.DefaultLogger.Info("MQTT Client already connected", "server", options.Servers()[0])
	}
	return nil
}

func (client *MqttClient) Disconnect() error {
	options := client.internalClient.OptionsReader()
	if client.internalClient.IsConnected() {
		log.DefaultLogger.Info("MQTT Client will disconnect", "server", options.Servers()[0])
		client.internalClient.Disconnect(100)
		log.DefaultLogger.Info("MQTT Client disconnected", "server", options.Servers()[0])
	} else {
		log.DefaultLogger.Info("MQTT Client not connected. No need to disconnect", "server", options.Servers()[0])
	}
	return nil
}

type TimedMessage struct {
	message         mqtt.Message
	acquisitionTime time.Time
}

/**
* Subscribe to the given topic
 */
func (client *MqttClient) Subscribe(sModel SubscriptionModel, callback func(TimedMessage)) error {
	tm, result := client.perTopicMutex.LoadOrStore(sModel.topic, &SubscriptionInfo{mut: &sync.Mutex{}, callbacks: make(map[SubscriptionModel]func(TimedMessage))})
	if !result {
		log.DefaultLogger.Info("This is the first subscription attempted", "topic", sModel.topic)
	}
	topicMutex := tm.(*SubscriptionInfo)
	topicMutex.mut.Lock()
	defer topicMutex.mut.Unlock()
	// Subscribe only if no grafana channel has subscribed
	if topicMutex.counter == 0 {
		subscriptionToken := client.internalClient.Subscribe(sModel.topic, client.qos, func(client mqtt.Client, mqttMessage mqtt.Message) {
			timedMessage := TimedMessage{
				message:         mqttMessage,
				acquisitionTime: time.Now(),
			}
			for _, callback := range topicMutex.callbacks {
				go func(c func(TimedMessage), m TimedMessage) {
					c(m)
				}(callback, timedMessage)
			}
		})
		if subscriptionToken.Wait() && subscriptionToken.Error() != nil {
			log.DefaultLogger.Error("Could not subscribe due to:" + subscriptionToken.Error().Error())
			return subscriptionToken.Error()
		}
	}
	if topicMutex.callbacks[sModel] == nil {
		topicMutex.callbacks[sModel] = callback
		topicMutex.counter = topicMutex.counter + 1
		log.DefaultLogger.Info("A new channel has subscribed", "topic", sModel.topic, "counter", topicMutex.counter)
	} else {
		log.DefaultLogger.Info("A channel has with the same info has subscribed", "info", sModel)
	}
	return nil
}

func (client *MqttClient) Unsubscribe(sModel SubscriptionModel) error {
	var errorResult error
	tm, result := client.perTopicMutex.Load(sModel.topic)
	if !result {
		return nil
	}
	topicMutex := tm.(*SubscriptionInfo)
	topicMutex.mut.Lock()
	defer topicMutex.mut.Unlock()
	log.DefaultLogger.Info("A channel will unsubscribe", "info", &sModel, "num of channels subsribed", topicMutex.counter)
	delete(topicMutex.callbacks, sModel)
	topicMutex.counter -= 1
	if topicMutex.counter == 0 {
		uToken := client.internalClient.Unsubscribe(sModel.topic)
		if uToken.Wait() && uToken.Error() != nil {
			log.DefaultLogger.Error("Could not unsubscribe due to:" + uToken.Error().Error())
			errorResult = uToken.Error()
		} else {
			log.DefaultLogger.Info("Successfully unsubscribed", "topic", sModel.topic)
		}
		client.perTopicMutex.Delete(sModel.topic)
	}
	return errorResult
}
