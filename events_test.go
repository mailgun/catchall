package catchall_test

import (
	"testing"

	"github.com/mailgun/catchall"
)

func BenchmarkEventBus(b *testing.B) {
	bus := catchall.SpawnEventPool()

	b.Run("GetEvent", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			e := bus.GetEvent()
			bus.RecycleEvent(e)
		}
	})
}
