package plugin

import (
	"encoding/json"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type FrameBuilder struct {
	data           map[string]*data.Field
	messageCounter int
}

func NewFrameBuilder() *FrameBuilder {
	return &FrameBuilder{data: make(map[string]*data.Field)}
}

func (fb *FrameBuilder) ToFrame() *data.Frame {
	frame := data.NewFrame("data")
	for _, field := range fb.data {
		frame.Fields = append(frame.Fields, field)
	}
	return frame
}

func (fb *FrameBuilder) AddData(acquisitionTime time.Time, jsonData map[string]json.RawMessage) {
	fb.messageCounter += 1
	fb.addTimeData(acquisitionTime)
	for key, value := range jsonData {
		log.DefaultLogger.Debug("json:", "key", key, "value", value)
		var parsedVal interface{}
		err := json.Unmarshal(value, &parsedVal)
		if err != nil {
			log.DefaultLogger.Error("Could not unmarshall json value:", value)
		}
		if fb.data[key] == nil {
			switch val := parsedVal.(type) {
			case string:
				fb.data[key] = data.NewField(key, nil, []string{val})
			case bool:
				fb.data[key] = data.NewField(key, nil, []bool{val})
			case float64:
				fb.data[key] = data.NewField(key, nil, []float64{val})
			case interface{}:
				fb.data[key] = data.NewField(key, nil, []json.RawMessage{value})
			case nil:

			}
		} else {
			fb.data[key].Extend(fb.messageCounter - fb.data[key].Len())
			fb.data[key].Set(fb.messageCounter-1, parsedVal)
		}
	}
}

func (fb *FrameBuilder) addTimeData(acquisitionTime time.Time) {
	if fb.data["Time"] == nil {
		fb.data["Time"] = data.NewField("Time", nil, []time.Time{acquisitionTime})
	} else {
		fb.data["Time"].Append(acquisitionTime)
	}
}

func (fb *FrameBuilder) AddRawData(acquisitionTime time.Time, raw []byte) {
	fb.messageCounter += 1
	fb.addTimeData(acquisitionTime)
	if fb.data["data"] == nil {
		fb.data["data"] = data.NewField("data", nil, []json.RawMessage{raw})
	} else {
		fb.data["data"].Extend(fb.messageCounter - fb.data["data"].Len())
		fb.data["data"].Set(fb.messageCounter-1, raw)
	}
}
