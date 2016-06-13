package main

import (
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
)

func init() {
	publishProviders["file"] = publishProviderFile{}
}

type publishProviderFile struct{}

func (p publishProviderFile) Write(targetURL *url.URL, buf io.Reader) error {
	f := strings.Replace(targetURL.String(), "file://", "", -1)
	d := path.Dir(f)

	if err := os.MkdirAll(d, 0755); err != nil {
		return err
	}

	data, err := ioutil.ReadAll(buf)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(f, data, 0644)
}
