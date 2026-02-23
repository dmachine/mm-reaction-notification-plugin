package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// Plugin implements the interface expected by the Mattermost server to provide
// thread reaction notifications.
type Plugin struct {
	plugin.MattermostPlugin

	// router handles HTTP requests routed to this plugin via ServeHTTP.
	router *mux.Router

	// configLock guards access to the configuration.
	configLock sync.RWMutex
}

// OnActivate is called when the plugin is loaded. It initializes the HTTP
// router and registers all API routes.
func (p *Plugin) OnActivate() error {
	p.router = mux.NewRouter()
	p.initAPI()
	return nil
}

// OnDeactivate is called when the plugin is unloaded or the server shuts down.
func (p *Plugin) OnDeactivate() error {
	return nil
}

// ServeHTTP routes all HTTP requests arriving at this plugin's URL prefix
// to the gorilla/mux router.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func main() {
	plugin.ClientMain(&Plugin{})
}
