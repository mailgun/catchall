package api

import "context"
import "github.com/256dpi/lungo"
import opts "go.mongodb.org/mongo-driver/mongo/options"
import "go.mongodb.org/mongo-driver/bson"

// Type representing the API state embedding a DB handle
// Could be improved with a proper external library and the use of indexes
type State struct {
    engine *lungo.Engine
    db lungo.IDatabase
}

// Type representing a domain, also used here as data model in the DB
// Could be improved by adding meta data from the events (created_at, etc...)
// or simply split the two usages, or even better using protobuf-like
// generated data structures instead
type Domain struct {
    Name string
    Deliveries int
    Bounced bool
}

// Enum type to define a domain type
type DomainType int
const (
    Unknown DomainType = iota
    CatchAll
    NotCatchAll
)

// Helper function to convert the type to string
func (this DomainType) String() string {
    ret := []string{"unknown", "catch-all", "not catch-all"}
    return ret[int(this)]
}

// Computes the type of the domain from its previous processed events
func (this *Domain) Type() DomainType {
    if this.Bounced {
        return NotCatchAll
    } else if this.Deliveries >= 1000 {
        return CatchAll
    }
    return Unknown
}

// Creates a new instance of the state with a default in-memory DB
func NewAPIState() (*State, error) {
    options := lungo.Options{
        Store: lungo.NewMemoryStore(),
    }
    client, engine, err := lungo.Open(context.TODO(), options)
    if err != nil {
        return nil, err
    }
    return &State{
        engine: engine,
        db: client.Database("cad"),
    }, nil
}

// Cleans up the DB handle
func (this *State) Clean() {
    this.engine.Close()
}

// Internal helper func to get a domain from the DB
func (this *State) getDomain(d *Domain) error {
    filter := bson.D{{"name", d.Name}}

    col := this.db.Collection("deliveries")
    err := col.FindOne(context.TODO(), filter).Decode(d)
    if err == lungo.ErrNoDocuments {
        *d = Domain{Name: d.Name}
        return nil
    }
    return err
}

// Internal helper func to update the DB when a domain is still determined as unknown
func (this *State) incDeliveries(d *Domain) error {
    filter := bson.D{{"name", d.Name}}
    update := bson.D{
        {"$inc", bson.D{{"deliveries", 1}}},
    }
    upsert := true
    options := opts.UpdateOptions{Upsert: &upsert}

    col := this.db.Collection("deliveries")
    _, err := col.UpdateOne(context.TODO(), filter, update, &options)
    return err
}

// Internal helper func to update the DB when a domain is determined as not catch-all
func (this *State) dropDeliveries(d *Domain) error {
    filter := bson.D{{"name", d.Name}}
    update := bson.D{
        {"$set", bson.D{{"bounced", true}}},
    }
    upsert := true
    options := opts.UpdateOptions{Upsert: &upsert}

    col := this.db.Collection("deliveries")
    _, err := col.UpdateOne(context.TODO(), filter, update, &options)
    return err
}
