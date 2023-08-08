package stages

import (
	"testing"

	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func Test_rdns_process(t *testing.T) {
	logger := util.TestFlowLogger(t)

	type fields struct {
		cfgs RDNSConfig
	}
	type args struct {
		labels    model.LabelSet
		extracted map[string]interface{}
	}
	field := "ip"
	defConf := RDNSConfig{
		Source: &field,
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			"Google DNS server",
			fields{defConf},
			args{
				labels: model.LabelSet{},
				extracted: map[string]interface{}{
					"ip": "8.8.8.8",
				},
			},
			map[string]interface{}{
				"ip":        "8.8.8.8",
				"hostnames": "google.dns",
			},
			false,
		},
		{
			"localhost",
			fields{defConf},
			args{
				labels: model.LabelSet{},
				extracted: map[string]interface{}{
					"ip": "127.0.0.1",
				},
			},
			map[string]interface{}{
				"ip":        "127.0.0.1",
				"hostnames": "localhost",
			},
			false,
		},
		{
			"unresolvable",
			fields{defConf},
			args{
				labels: model.LabelSet{},
				extracted: map[string]interface{}{
					"ip": "1.2.3.4",
				},
			},
			map[string]interface{}{
				"ip": "1.2.3.4",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &reverseDNSStage{
				logger: logger,
				cfgs:   tt.fields.cfgs,
			}
			g.process(tt.args.labels, tt.args.extracted)
			require.Equal(t, tt.expected, tt.args.extracted)
		})
	}
}
