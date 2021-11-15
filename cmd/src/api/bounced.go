package api

import "io"

// Process a bounced event
func (this *State) Bounced(w io.Writer, vars map[string]string) error {
    d := Domain{Name: vars["domain"]}
    return this.dropDeliveries(&d)
}
