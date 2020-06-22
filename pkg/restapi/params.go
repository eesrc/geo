package restapi

import "strings"

type ACMEParameters struct {
	Enabled   bool   `param:"desc=Enable ACME Certificates (aka Let's Encrypt) for host;default=false"`
	Hosts     string `param:"desc=ACME host names;default=geo.exploratory.engineering,geo.nbiot.engineering,api.geo.telenor.io"`
	SecretDir string `param:"desc=ACME secrets directory;default=/var/geo/autocert"`
}

//HostList returns the list of hosts
func (p *ACMEParameters) HostList() []string {
	return strings.Split(p.Hosts, ",")
}

type RestAPIParams struct {
	Endpoint    string `param:"desc=Listen address for HTTP server;default=localhost:8080"`
	TLSKeyFile  string `param:"desc=TLS key file;file"`
	TLSCertFile string `param:"desc=TLS certificate file;file"`
	ACME        ACMEParameters
	AccessLog   string `param:"desc=Access log file name;default=access_log"`
}
