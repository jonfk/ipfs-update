package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	api "github.com/ipfs/go-ipfs-api"
	. "github.com/whyrusleeping/stump"
)

func httpFetch(url string) (io.ReadCloser, error) {
	VLog("fetching url: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http.Get error: %s", err)
	}

	if resp.StatusCode >= 400 {
		Error("fetching resource: %s", resp.Status)
		mes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading error body: %s", err)
		}

		return nil, fmt.Errorf("%s: %s", resp.Status, string(mes))
	}

	return resp.Body, nil
}

func Fetch(ipfspath string) (io.ReadCloser, error) {
	VLog("  - fetching %q", ipfspath)
	sh := api.NewShell("http://localhost:5001")
	if sh.IsUp() {
		VLog("  - using local ipfs daemon for transfer")
		return sh.Cat(ipfspath)
	}

	return httpFetch(gateway + ipfspath)
}

// This function is needed because os.Rename doesnt work across filesystem
// boundaries.
func CopyTo(src, dest string) error {
	VLog("  - copying %s to %s", src, dest)
	fi, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fi.Close()

	trgt, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer trgt.Close()

	_, err = io.Copy(trgt, fi)
	return err
}

func Move(src, dest string) error {
	err := CopyTo(src, dest)
	if err != nil {
		return err
	}

	return os.Remove(src)
}

func ipfsDir() string {
	def := filepath.Join(os.Getenv("HOME"), ".ipfs")

	ipfs_path := os.Getenv("IPFS_PATH")
	if ipfs_path != "" {
		def = ipfs_path
	}

	return def
}

func hasDaemonRunning() bool {
	shell := api.NewShell("localhost:5001")
	return shell.IsUp()
}