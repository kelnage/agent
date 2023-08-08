package stages

import (
	"errors"
	"net"
	"reflect"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/miekg/dns"
	"github.com/prometheus/common/model"
)

var (
	ErrEmptyRDNSLookupConfig      = errors.New("reverse_dns stage config cannot be empty")
	ErrEmptySourceRDNSStageConfig = errors.New("source cannot be empty")
)

// RDNSConfig represents Reverse DNS stage config
type RDNSConfig struct {
	Source *string `river:"source,attr"`
}

func validateRDNSConfig(c RDNSConfig) error {
	if c.Source != nil && *c.Source == "" {
		return ErrEmptySourceRDNSStageConfig
	}
	return nil
}

func newRDNSStage(logger log.Logger, config RDNSConfig) (Stage, error) {
	err := validateRDNSConfig(config)
	if err != nil {
		return nil, err
	}

	return &reverseDNSStage{
		logger: logger,
		cfgs:   config,
	}, nil
}

type reverseDNSStage struct {
	logger log.Logger
	cfgs   RDNSConfig
}

// Run implements Stage
func (g *reverseDNSStage) Run(in chan Entry) chan Entry {
	out := make(chan Entry)
	go func() {
		defer close(out)
		defer g.close()
		for e := range in {
			g.process(e.Labels, e.Extracted)
			out <- e
		}
	}()
	return out
}

// Name implements Stage
func (g *reverseDNSStage) Name() string {
	return StageTypeReverseDNS
}

func (g *reverseDNSStage) process(_ model.LabelSet, extracted map[string]interface{}) {
	var ip net.IP
	if g.cfgs.Source != nil {
		if _, ok := extracted[*g.cfgs.Source]; !ok {
			if Debug {
				level.Debug(g.logger).Log("msg", "source does not exist in the set of extracted values", "source", *g.cfgs.Source)
			}
			return
		}

		value, err := getString(extracted[*g.cfgs.Source])
		if err != nil {
			if Debug {
				level.Debug(g.logger).Log("msg", "failed to convert source value to string", "source", *g.cfgs.Source, "err", err, "type", reflect.TypeOf(extracted[*g.cfgs.Source]))
			}
			return
		}
		ip = net.ParseIP(value)
		if ip == nil {
			level.Error(g.logger).Log("msg", "source is not an ip", "source", value)
			return
		}
	}
	// TODO: allow configuration to use a specific resolver
	names, err := net.LookupAddr(ip.String())
	if err != nil {
		level.Error(g.logger).Log("msg", "dns lookup failed", "source", ip.String())
		return
	}
	if len(names) > 0 {
		hostnames := ""
		for i, name := range names {
			hostnames += normaliseHost(name)
			if i < len(names)-1 {
				hostnames += ";"
			}
		}
		extracted["hostnames"] = hostnames
	}
}

func (d *reverseDNSStage) close() {
	// NOP?
}

func normaliseHost(ptr string) string {
	host := ""
	labels := dns.SplitDomainName(ptr)
	if len(labels) > 0 {
		for j := len(labels) - 1; j > 0; j-- {
			host += labels[j] + "."
		}
		host += labels[0]
	}
	return host
}
