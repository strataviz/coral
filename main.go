package main

import "stvz.io/coral/pkg/cmd"

func main() {
	root := cmd.NewRoot()
	err := root.Execute()
	if err != nil {
		panic(err)
	}
}
