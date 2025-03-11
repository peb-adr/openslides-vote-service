package vote_test

import (
	"context"

	"github.com/peb-adr/openslides-go/datastore/dskey"
)

type StubGetter struct {
	data      map[dskey.Key][]byte
	err       error
	requested map[dskey.Key]bool
}

func (g *StubGetter) Get(ctx context.Context, keys ...dskey.Key) (map[dskey.Key][]byte, error) {
	if g.err != nil {
		return nil, g.err
	}
	if g.requested == nil {
		g.requested = make(map[dskey.Key]bool)
	}

	out := make(map[dskey.Key][]byte, len(keys))
	for _, k := range keys {
		out[k] = g.data[k]
		g.requested[k] = true
	}
	return out, nil
}

func (g *StubGetter) Update(context.Context, func(map[dskey.Key][]byte, error)) {}
