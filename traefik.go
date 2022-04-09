package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// HTTPConfiguration contains all the HTTP configuration parameters.
type HTTPConfiguration struct {
	Routers           map[string]*Router           `json:"routers,omitempty" toml:"routers,omitempty" yaml:"routers,omitempty" export:"true"`
	Services          map[string]*Service          `json:"services,omitempty" toml:"services,omitempty" yaml:"services,omitempty" export:"true"`
	Middlewares       map[string]*Middleware       `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	ServersTransports map[string]*ServersTransport `json:"serversTransports,omitempty" toml:"serversTransports,omitempty" yaml:"serversTransports,omitempty" label:"-" export:"true"`
}

// Service holds a service configuration (can only be of one type at the same time).
type Service struct {
	LoadBalancer *ServersLoadBalancer `json:"loadBalancer,omitempty" toml:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty" export:"true"`
}

// Router holds the router configuration.
type Router struct {
	EntryPoints []string         `json:"entryPoints,omitempty" toml:"entryPoints,omitempty" yaml:"entryPoints,omitempty" export:"true"`
	Middlewares []string         `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
	Service     string           `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty" export:"true"`
	Rule        string           `json:"rule,omitempty" toml:"rule,omitempty" yaml:"rule,omitempty"`
	Priority    int              `json:"priority,omitempty" toml:"priority,omitempty,omitzero" yaml:"priority,omitempty" export:"true"`
	TLS         *RouterTLSConfig `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// RouterTLSConfig holds the TLS configuration for a router.
type RouterTLSConfig struct {
	Options      string   `json:"options,omitempty" toml:"options,omitempty" yaml:"options,omitempty" export:"true"`
	CertResolver string   `json:"certResolver,omitempty" toml:"certResolver,omitempty" yaml:"certResolver,omitempty" export:"true"`
	Domains      []Domain `json:"domains,omitempty" toml:"domains,omitempty" yaml:"domains,omitempty" export:"true"`
}

// ServersLoadBalancer holds the ServersLoadBalancer configuration.
type ServersLoadBalancer struct {
	Servers          []Server `json:"servers,omitempty" toml:"servers,omitempty" yaml:"servers,omitempty" label-slice-as-struct:"server" export:"true"`
	ServersTransport string   `json:"serversTransport,omitempty" toml:"serversTransport,omitempty" yaml:"serversTransport,omitempty" export:"true"`
}

// Server holds the server configuration.
type Server struct {
	URL string `json:"url,omitempty" toml:"url,omitempty" yaml:"url,omitempty" label:"-"`
}

// ServersTransport options to configure communication between Traefik and the servers.
type ServersTransport struct {
	InsecureSkipVerify bool `description:"Disable SSL certificate verification." json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
}

// Domain holds a domain name with SANs.
type Domain struct {
	Main string   `description:"Default subject name." json:"main,omitempty" toml:"main,omitempty" yaml:"main,omitempty"`
	SANs []string `description:"Subject alternative names." json:"sans,omitempty" toml:"sans,omitempty" yaml:"sans,omitempty"`
}

// Middleware holds the Middleware configuration.
type Middleware struct {
	RedirectRegex *RedirectRegex `json:"redirectRegex,omitempty" toml:"redirectRegex,omitempty" yaml:"redirectRegex,omitempty" export:"true"`
}

// RedirectRegex holds the redirection configuration.
type RedirectRegex struct {
	Regex       string `json:"regex,omitempty" toml:"regex,omitempty" yaml:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty" toml:"replacement,omitempty" yaml:"replacement,omitempty"`
	Permanent   bool   `json:"permanent,omitempty" toml:"permanent,omitempty" yaml:"permanent,omitempty" export:"true"`
}

type TraefikHandler struct {
	config HTTPConfiguration
}

func NewTraefikHandler(domain string, bookmarks []*Bookmark) *TraefikHandler {
	return &TraefikHandler{
		config: ToHTTPConfiguration(domain, bookmarks),
	}
}

func ToHTTPConfiguration(domain string, bookmarks []*Bookmark) HTTPConfiguration {
	routers := make(map[string]*Router)
	services := make(map[string]*Service)
	middlewares := make(map[string]*Middleware)
	for _, bookmark := range bookmarks {
		for k, router := range ToRouter(domain, bookmark) {
			routers[k] = router
		}
		for k, service := range ToService(bookmark) {
			services[k] = service
		}
		for k, middleware := range ToMiddleware(domain, bookmark) {
			middlewares[k] = middleware
		}
	}
	return HTTPConfiguration{
		Routers:     routers,
		Services:    services,
		Middlewares: middlewares,
		ServersTransports: map[string]*ServersTransport{
			"tinybp-insecure-skip-verify": &ServersTransport{
				InsecureSkipVerify: true,
			},
		},
	}
}

func ToRouter(domain string, bookmark *Bookmark) map[string]*Router {
	entrypoints := []string{"http"}
	if intEntryPts, ok := bookmark.LinkerConfig["entryPoints"]; ok {
		if newEntryPoints, ok := intEntryPts.([]interface{}); ok {
			entrypoints = make([]string, 0)
			for _, e := range newEntryPoints {
				entrypoints = append(entrypoints, e.(string))
			}

		}
	}
	middlewares := []string{fmt.Sprintf("redirect-tinybp-%s", bookmark.Name)}
	if bookmark.Proxify {
		middlewares = make([]string, 0)
	}
	name := fmt.Sprintf("tinybp-%s", bookmark.Name)
	rule := fmt.Sprintf("Host(`%s.%s`)", bookmark.Name, domain)
	routers := make(map[string]*Router)
	routers[name] = &Router{
		EntryPoints: entrypoints,
		Service:     name,
		Middlewares: middlewares,
		Rule:        rule,
	}
	if _, ok := bookmark.LinkerConfig["enableTls"]; ok {
		routers[name+"-tls"] = &Router{
			EntryPoints: entrypoints,
			Service:     name,
			Rule:        rule,
			Middlewares: middlewares,
			TLS:         &RouterTLSConfig{},
		}
	}
	return routers
}

func ToService(bookmark *Bookmark) map[string]*Service {
	transport := ""
	if bookmark.InsecureSkipVerify {
		transport = "tinybp-insecure-skip-verify"
	}
	name := fmt.Sprintf("tinybp-%s", bookmark.Name)
	return map[string]*Service{
		name: &Service{
			LoadBalancer: &ServersLoadBalancer{
				ServersTransport: transport,
				Servers: []Server{
					{
						URL: bookmark.Url,
					},
				},
			},
		},
	}
}

func ToMiddleware(domain string, bookmark *Bookmark) map[string]*Middleware {
	if bookmark.Proxify {
		return map[string]*Middleware{}
	}
	regex := fmt.Sprintf("http(s)?://%s.%s(/.*)", bookmark.Name, domain)
	url := strings.TrimSuffix(bookmark.Url, "/")
	return map[string]*Middleware{
		fmt.Sprintf("redirect-tinybp-%s", bookmark.Name): &Middleware{
			RedirectRegex: &RedirectRegex{
				Regex:       regex,
				Replacement: fmt.Sprintf("%s${2}", url),
				Permanent:   false,
			},
		},
	}
}

func (h *TraefikHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(struct {
		Http HTTPConfiguration `json:"http"`
	}{
		Http: h.config,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
