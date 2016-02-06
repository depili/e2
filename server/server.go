package server

import (
    "github.com/qmsk/e2/client"
)

type Options struct {

}

func (options Options) Server(clientClient *client.Client) (*Server, error) {
    server := &Server{
        client:     clientClient,
    }

    return server, nil
}

type Server struct {
    client      *client.Client
}

func (server *Server) Index(name string) (apiResource, error) {
    switch name {
    case "":
        index := Index{}

        return &index, index.load(server.client)

    case "status":
        status := Status{
            client: server.client,
        }

        return &status, nil

    case "sources":
        sources := Sources{}

        return &sources, sources.load(server.client)

    case "screens":
        screens := Screens{
            client:     server.client,
        }

        return &screens, screens.load(server.client)

    case "auxes":
        auxes := Auxes{}

        return &auxes, auxes.load(server.client)

    case "presets":
        presets := Presets{
            client:     server.client,
        }

        return &presets, presets.load(server.client)

    default:
        return nil, nil
    }
}
