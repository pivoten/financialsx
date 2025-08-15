//go:build darwin
// +build darwin

package main

/*
#cgo darwin LDFLAGS: -framework UniformTypeIdentifiers -framework CoreServices
*/
import "C"