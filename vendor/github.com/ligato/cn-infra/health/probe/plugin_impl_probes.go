// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package probe

import (
	"encoding/json"
	"net/http"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/health/statuscheck/model/status"
	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
	"github.com/unrolled/render"
)

const (
	// Default port used for http and probing
	defaultPort               = "9191"
	livenessProbePath  string = "/liveness"  // liveness probe URL
	readinessProbePath string = "/readiness" // readiness probe URL
)

// PluginID used in the Agent Core flavors
const PluginID core.PluginName = "Probe"

var (
	httpPort string
)

// init is here only for parsing program arguments
func init() {
	flag.StringVar(&httpPort, "http-probe-port", defaultPort,
		"Listen port for the Agent's HTTPProbe server.")
}

// Plugin struct holds all plugin-related data
type Plugin struct {
	StatusCheck statuscheck.AgentStatusReader
	Log         logging.PluginLogger

	HTTP *rest.Plugin

	customProbe bool
}

// Init is the plugin entry point called by the Agent Core
func (p *Plugin) Init() (err error) {
	// Start Init() and AfterInit() for new probe in case the port is different from agent http

	if p.HTTP.HTTPport != httpPort {
		p.Log.Warnf("Custom port: %v", httpPort)
		p.HTTP = &rest.Plugin{
			Log:      p.Log,
			HTTPport: httpPort,
		}
		err := p.HTTP.Init()
		if err != nil {
			return err
		}
		err = p.HTTP.AfterInit()
		if err != nil {
			return err
		}

		p.customProbe = true
	}
	return nil
}

// AfterInit is called by the Agent Core after all plugins have been initialized.
func (p *Plugin) AfterInit() error {
	if p.HTTP != nil {
		p.Log.Warnf("Starting health probes on port %v", p.HTTP.HTTPport)
		p.HTTP.RegisterHTTPHandler(livenessProbePath, p.livenessProbeHandler, "GET")
		p.HTTP.RegisterHTTPHandler(readinessProbePath, p.readinessProbeHandler, "GET")
	}

	return nil
}

// Close is called by the Agent Core when it's time to clean up the plugin
func (p *Plugin) Close() error {
	if p.customProbe {
		_, err := safeclose.CloseAll(p.HTTP)
		return err
	}

	return nil
}

// readinessProbeHandler handles k8s readiness probe.
func (p *Plugin) readinessProbeHandler(formatter *render.Render) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		stat := p.StatusCheck.GetAgentStatus()
		statJSON, _ := json.Marshal(stat)
		if stat.State == status.OperationalState_OK {
			w.WriteHeader(http.StatusOK)
			w.Write(statJSON)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(statJSON)
		}
	}
}

// livenessProbeHandler handles k8s liveness probe.
func (p *Plugin) livenessProbeHandler(formatter *render.Render) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		stat := p.StatusCheck.GetAgentStatus()
		statJSON, _ := json.Marshal(p.StatusCheck.GetAgentStatus())

		if stat.State == status.OperationalState_INIT || stat.State == status.OperationalState_OK {
			w.WriteHeader(http.StatusOK)
			w.Write(statJSON)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(statJSON)
		}
	}
}
