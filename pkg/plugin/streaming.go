package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/reactivex/rxgo/v2"
)

type SubscriptionModel struct {
	topic         string
	includeSchema bool
	useInterval   bool
	interval      int
}

// SubscribeStream implements backend.StreamHandler
func (d *Datasource) SubscribeStream(ctx context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	logger.Info("Subscribing", "channel", req.Path)
	status := backend.SubscribeStreamStatusOK
	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// PublishStream called when a user tries to publish to a plugin/datasource
// managed channel path. Here plugin can check publish permissions and
// modify publication data if required.
func (*Datasource) PublishStream(context.Context, *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	panic("unimplemented")
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (d *Datasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream called", "path", req.Path)

	sm, err := NewSubscriptionModel(req.Path)
	if err != nil {
		return err
	}
	// Create the input channel
	ch := make(chan rxgo.Item)
	err = d.mqttClient.Subscribe(sm, func(m TimedMessage) {
		ch <- rxgo.Of(m)
	})
	if err != nil {
		return err
	}
	defer func() {
		log.DefaultLogger.Info("Unsubscribing", "path", req.Path)
		err = d.mqttClient.Unsubscribe(sm)
	}()

	dataStream := rxgo.FromChannel(ch)
	if sm.useInterval {
		dataStream = dataStream.BufferWithTime(rxgo.WithDuration(time.Duration(sm.interval * 1000000)))
	}
	dataStream = dataStream.Map(func(_ context.Context, item interface{}) (interface{}, error) {
		//log.DefaultLogger.Debug("Received message")
		var messages []interface{}
		_, ok := item.(TimedMessage)
		if !ok {
			//log.DefaultLogger.Debug("buffered message")
			messages, ok = item.([]interface{})
			if !ok {
				log.DefaultLogger.Error("error multiple")
				return item, fmt.Errorf("unknown item")
			}
		} else {
			messages = append(messages, item)
		}
		fb := NewFrameBuilder()
		for _, m := range messages {
			mqttMessage := m.(TimedMessage)
			log.DefaultLogger.Debug("Received message", "data", string(mqttMessage.message.Payload()))
			var jsonData map[string]json.RawMessage
			err := json.Unmarshal(mqttMessage.message.Payload(), &jsonData)
			if err != nil || !sm.includeSchema {
				fb.AddRawData(mqttMessage.acquisitionTime, mqttMessage.message.Payload())
			} else {
				fb.AddData(mqttMessage.acquisitionTime, jsonData)
			}
		}
		sender.SendFrame(fb.ToFrame(), data.IncludeAll)
		return item, nil
	})
	outputStream := dataStream.Observe()
	defer close(ch)
	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Info("Context canceled for channel", "path", req.Path)
			return nil
		case <-outputStream:
			//
		}
	}
}
