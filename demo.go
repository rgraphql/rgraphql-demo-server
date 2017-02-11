package main

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"
)

var schemaAst string = `
type Person {
	name: String
	steps: Int
}

type RootQuery {
	allPeople: [Person]
}

schema {
	query: RootQuery
}
`

type RootQueryResolver struct{}

func (r *RootQueryResolver) AllPeople() []*PersonResolver {
	return []*PersonResolver{
		{name: "Tom", steps: 50},
		{name: "Jerry", steps: 1000},
	}
}

type PersonResolver struct {
	name  string
	steps int
}

func (r *PersonResolver) Name() string {
	return r.name
}

func (r *PersonResolver) Steps(ctx context.Context, output chan<- int) error {
	done := ctx.Done()
	for {
		log.WithField("steps", r.steps).Debug("Sending steps increment")
		output <- r.steps
		r.steps++

		select {
		case <-done:
			return nil
		case <-time.After(time.Duration(1) * time.Second):
		}
	}
}
