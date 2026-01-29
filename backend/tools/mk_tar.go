package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"os"
)

func main() {
	// Target path
	path := "../deployments/tars/os/ubuntu/amd64/jammy/go1.21.linux-amd64.tar.gz"
	
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add a dummy file inside
	body := "go version go1.21.0 linux/amd64"
	hdr := &tar.Header{
		Name: "go/VERSION",
		Mode: 0600,
		Size: int64(len(body)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		panic(err)
	}
	if _, err := tw.Write([]byte(body)); err != nil {
		panic(err)
	}

	fmt.Println("Valid tar.gz created at", path)
}
