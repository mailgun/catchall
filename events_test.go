package catchall_test

import (
	"log"
	"net/http"
	"testing"

	"github.com/mailgun/catchall"
)

func BenchmarkEventBus(b *testing.B) {
	bus := catchall.SpawnEventPool()
	client := &http.Client{}

	b.Run("GetEvent", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			e := bus.GetEvent()
			uri_root := "http://localhost:8080/events/" + e.Domain + "/"
			action := "delivered"
			if e.Type == catchall.TypeBounced {
				action = "bounced"
			}
			req, err := http.NewRequest(http.MethodPut, uri_root+action, nil)
			if err != nil {
				log.Fatal(err)
			}
			resp, err := client.Do(req)
			resp.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			bus.RecycleEvent(e)
		}
	})
}
