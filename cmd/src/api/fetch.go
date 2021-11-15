package api

import "io"

// Fetchs the current domain type
func (this *State) FetchDomain(w io.Writer, vars map[string]string) error {
    d := Domain{Name: vars["domain"]}
    err := this.getDomain(&d)
    if err != nil {
        return err
    }
    io.WriteString(w, d.Type().String())
    return nil
}
