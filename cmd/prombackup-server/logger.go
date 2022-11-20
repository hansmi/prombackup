package main

type Logger interface {
	Print(...any)
	Printf(string, ...any)
}
