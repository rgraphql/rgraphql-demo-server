package main

import (
	"context"
	"time"
)

var schemaAst string = `
type Person {
	name: String
	steps: Int
	favoriteMonths: [String]
	parents: [Person]
	wordsSaid: [String]
	stuff: [[[[String]]]]
}

type RootQuery {
	allPeople(initialSteps: Int): [Person]
	person(name: String): Person
}

schema {
	query: RootQuery
}
`

var jerry = &PersonResolver{
	name:  "Jerry",
	steps: 1000,
}

var mary = &PersonResolver{
	name:  "Mary",
	steps: 2000,
}

var parents = []*PersonResolver{
	jerry,
	mary,
}

func (p *PersonResolver) Stuff() [][][][]string {
	return nil
}

type RootQueryResolver struct{}

func (r *RootQueryResolver) AllPeople(args *struct{ InitialSteps int }) []*PersonResolver {
	if args == nil {
		return nil
	}

	return []*PersonResolver{
		{name: "Tommy", steps: args.InitialSteps, parents: parents},
		{name: "Billy", steps: args.InitialSteps * 2, parents: parents},
	}
}

func (r *RootQueryResolver) Person(args *struct{ Name string }) *PersonResolver {
	if args == nil || args.Name == "" {
		return nil
	}

	return &PersonResolver{
		name:    args.Name,
		parents: parents,
	}
}

type PersonResolver struct {
	name    string
	steps   int
	parents []*PersonResolver
}

func (r *PersonResolver) Name() string {
	return r.name
}

func (r *PersonResolver) Parents() []*PersonResolver {
	return r.parents
}

func (r *PersonResolver) FavoriteMonths(ctx context.Context) <-chan string {
	ch := make(chan string)
	result := []string{
		"January",
		"March",
		"August",
		"December",
	}
	go func() {
		for _, res := range result {
			select {
			case ch <- res:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
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

func (r *PersonResolver) WordsSaid(ctx context.Context, output chan<- (<-chan string)) error {
	return nil
}
