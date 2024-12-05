package main

import "github.com/go-playground/webhooks/v6/gitlab"

func main() {
	gitlab.New(gitlab.Options.Secret("12345"))
}
