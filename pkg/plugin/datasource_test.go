package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/assert"
)

func TestQueryData(t *testing.T) {
	ds := Datasource{}

	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{RefID: "A"},
			},
		},
	)
	if err != nil {
		t.Error(err)
	}

	if len(resp.Responses) != 1 {
		t.Fatal("QueryData must return a response")
	}
	assert.Equal(t, resp.Responses["A"].Status, backend.Status(backend.StatusBadRequest))
}

func TestSingleQuery(t *testing.T) {
	testCases := []struct {
		qm    QueryModel
		valid bool
	}{
		{QueryModel{Topic: "   "}, false},
		{QueryModel{Topic: "top"}, true},
		{QueryModel{Topic: "08:08"}, true},
	}

	ctx := context.Background()
	ds := Datasource{}
	for _, tc := range testCases {
		jsonData, err := json.Marshal(QueryModel{Topic: tc.qm.Topic})
		if err != nil {
			t.Error(err)
		}
		resp, err := ds.QueryData(
			ctx,
			&backend.QueryDataRequest{
				Queries: []backend.DataQuery{
					{RefID: "A", JSON: jsonData},
				},
				PluginContext: backend.PluginContext{DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{UID: "1"}},
			},
		)
		if err != nil {
			t.Error(err)
		}
		if tc.valid {
			assert.Equal(t, resp.Responses["A"].Frames[0].Meta.Channel, fmt.Sprintf("ds/1/topic=%s.useInterval=false.includeSchema=false.interval=0s", tc.qm.Topic))
		} else {
			assert.False(t, tc.valid)
		}
	}
}
