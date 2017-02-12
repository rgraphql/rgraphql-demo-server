package main

import (
	"context"
	"time"
)

var schemaAst string = `
type Person {
	name: String
	steps: Int
}

type RootQuery {
	allPeople: [Person]
	person(name: String): Person
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

func (r *RootQueryResolver) Person(args *struct{ Name string }) *PersonResolver {
	if args == nil || args.Name == "" {
		return nil
	}

	return &PersonResolver{
		name: args.Name,
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
		output <- r.steps
		r.steps++

		select {
		case <-done:
			return nil
		case <-time.After(time.Duration(1) * time.Second):
		}
	}
}
