package main

import "os"

func writeFile(name string, data []byte) error {
	return os.WriteFile(name, data, 0o644)
}
