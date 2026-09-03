package main

import "fmt"

type GreetService struct{}

func (g *GreetService) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's Showtime!", name)
}

func (g *GreetService) GreetMany(names []string) []string {
	greetings := make([]string, len(names))
	for i, name := range names {
		greetings[i] = fmt.Sprintf("Hello %s!", name)
	}
	return greetings
}
