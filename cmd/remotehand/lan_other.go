//go:build !windows

package main

func physicalLAN(int) bool { return true }
