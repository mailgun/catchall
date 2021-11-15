package main

import "log"
import "cad/src/handler"
import "cad/src/api"

// Init the API state and register the routes through the handler
func main() {
    state, err := api.NewAPIState()
    if err != nil {
        log.Fatal(err)
    }
    defer state.Clean()

    hdl := handler.NewHandler()
    hdl.RegisterEvent(&handler.Route{
        Method: "PUT",
        Path: "/events/{domain}/delivered",
        Name: "delivered",
        Func: state.Delivered,
    })
    hdl.RegisterEvent(&handler.Route{
        Method: "PUT",
        Path: "/events/{domain}/bounced",
        Name: "bounced",
        Func: state.Bounced,
    })
    hdl.RegisterRoute(&handler.Route{
        Method: "GET",
        Path: "/domains/{domain}",
        Func: state.FetchDomain,
    })
    err = hdl.Listen()
    if err != nil {
        log.Fatal(err)
    }
}
