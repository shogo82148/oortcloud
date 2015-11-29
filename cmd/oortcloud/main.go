package main

import (
	"errors"
	"flag"
	"net"
	"net/http"
	_ "net/http/pprof"

	log "github.com/Sirupsen/logrus"
	"github.com/shogo82148/oortcloud"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "oorcloud.yml", "path to configure file")
	flag.StringVar(&configPath, "config", "oorcloud.yml", "path to configure file")
	flag.Parse()

	config, err := LoadConfig(configPath)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Fatal("error while reading configuration file")
	}

	notifier := oortcloud.NewHTTPNotifier(config.API.Callback)
	notifier.Client.Transport.(*http.Transport).MaxIdleConnsPerHost = config.API.MaxIdleConnsPerHost

	// start websocket server
	if ws := config.Websocket; ws != nil {
		connector := oortcloud.NewWebSocketConnector(notifier, ws.Binary)
		connectorListener, err := NewListener(ws.Host, ws.Port, ws.Sock)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Fatal("listening Websocket port failed")
		}
		connectorServer := &http.Server{
			Handler: connector,
		}
		go func() {
			err := connectorServer.Serve(connectorListener)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err.Error(),
				}).Fatal("starting Websocket server failed")
			}
		}()
	}

	// start api server
	notifierListener, err := NewListener(config.API.Host, config.API.Port, config.API.Sock)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Fatal("listening API port failed")
	}
	notifierServer := &http.Server{
		Handler: notifier,
	}
	err = notifierServer.Serve(notifierListener)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Fatal("starting API server failed")
	}
}

func NewListener(host, port, sock string) (net.Listener, error) {
	if sock != "" {
		return net.Listen("unix", sock)
	}
	if port != "" {
		return net.Listen("tcp", net.JoinHostPort(host, port))
	}

	return nil, errors.New("missing listen port or unix domain socket")
}
