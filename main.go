package main

import (
	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/server"
	"github.com/NYTimes/video-captions-api/service"
)

func main() {
	var cfg service.Config
	config.LoadJSONFile("./config.json", &cfg)

	server.Init("video-captions-api", cfg.Server)

	err := server.Register(service.NewSimpleService(&cfg))
	if err != nil {
		server.Log.Fatal("Unable to register service: ", err)
	}

	err = server.Run()
	if err != nil {
		server.Log.Fatal("Server encountered a fatal error: ", err)
	}

}
