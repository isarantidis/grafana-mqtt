package plugin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

var logger = log.DefaultLogger

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces- only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.StreamHandler         = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	logger.Info("Creating datasource")
	client, error := createMqttClient(0, settings)
	if error != nil {
		return nil, error
	}
	error = client.Connect()
	if error != nil {
		return nil, error
	}
	return &Datasource{
		settings:   settings,
		mqttClient: client,
	}, nil
}

func createMqttClient(qos uint8, instanceSettings backend.DataSourceInstanceSettings) (*MqttClient, error) {
	logger.Info("Creating Mqtt client", "settings", instanceSettings)
	var datasourceSettings DatasourceSettings
	if err := parseDatasourceConfig(instanceSettings.JSONData, instanceSettings.DecryptedSecureJSONData, &datasourceSettings); err != nil {
		return nil, fmt.Errorf("error parsing settings: %v", err)
	}

	var options = mqtt.NewClientOptions()
	options.SetUsername(datasourceSettings.Username)
	options.SetPassword(datasourceSettings.Password)
	options.SetClientID(datasourceSettings.ClientId)
	options.AddBroker(datasourceSettings.BrokerUrl)
	options.SetConnectionAttemptHandler(func(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
		log.DefaultLogger.Info("Attempt to connect", "url", broker.String())
		return tlsCfg
	})
	options.SetOnConnectHandler(func(c mqtt.Client) {
		options := c.OptionsReader()
		log.DefaultLogger.Info("Connected", "url", options.Servers()[0])
	})

	client := mqtt.NewClient(options)
	return NewMQTTManager(uint8(datasourceSettings.QoS), client), nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	settings   backend.DataSourceInstanceSettings
	mqttClient *MqttClient
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	d.mqttClient.Disconnect()
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type QueryModel struct {
	Topic         string `json:"topic,omitempty"`
	UseInterval   bool   `json:"useInterval,omitempty"`
	IncludeSchema bool   `json:"includeSchema,omitempty"`
}

var regex *regexp.Regexp

func init() {
	var err error
	regex, err = regexp.Compile(`[^\S-]`)
	if err != nil {
		panic(err.Error())
	}
}

func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	// Unmarshal the JSON into our queryModel.
	var qm QueryModel
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}

	if regex.Match([]byte(qm.Topic)) {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf(""))
	}

	// create data frame response.
	// For an overview on data frames and how grafana handles them:
	// https://grafana.com/docs/grafana/latest/developers/plugins/data-frames/
	frame := data.NewFrame("")

	channelPath := fmt.Sprintf("topic=%s.useInterval=%v.includeSchema=%v.interval=%v", qm.Topic, qm.UseInterval, qm.IncludeSchema, query.Interval)
	frame.SetMeta(&data.FrameMeta{
		Channel: path.Join("ds", pCtx.DataSourceInstanceSettings.UID, channelPath),
	})

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)
	return response
}

type DatasourceSettings struct {
	BrokerUrl string `json:"brokerUrl,omitempty"`
	ClientId  string `json:"clientId,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	QoS       uint8  `json:"qos"`
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("Health check1")
	var status = backend.HealthStatusOk
	var message = "Data source is working"

	err := d.mqttClient.Connect()
	// Connect to the MQTT broker
	if err != nil {
		log.DefaultLogger.Error("Error connecting to MQTT broker: %v", err.Error())
		status = backend.HealthStatusError
		message = err.Error()
	}
	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

func parseDatasourceConfig(jsonData []byte, secureJsonMap map[string]string, datasourceSettings *DatasourceSettings) error {
	if err := json.Unmarshal(jsonData, datasourceSettings); err != nil {
		log.DefaultLogger.Error(err.Error())
		return errors.New(err.Error())
	}
	for key, value := range secureJsonMap {
		switch key {
		case "password":
			datasourceSettings.Password = value
		}
	}
	return nil
}
