package api

import "io"

// Process a delivered event
func (this *State) Delivered(w io.Writer, vars map[string]string) error {
    d := Domain{Name: vars["domain"]}
    return this.incDeliveries(&d)
}
