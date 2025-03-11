package vote

import (
	"fmt"

	"github.com/peb-adr/openslides-go/datastore"
	"github.com/peb-adr/openslides-go/datastore/cache"
	"github.com/peb-adr/openslides-go/datastore/flow"
	"github.com/peb-adr/openslides-go/environment"
)

// Flow initializes a cached connection to postgres.
func Flow(lookup environment.Environmenter, messageBus flow.Updater) (flow.Flow, error) {
	postgres, err := datastore.NewFlowPostgres(lookup, messageBus)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}

	cache := cache.New(postgres)

	return cache, nil
}
