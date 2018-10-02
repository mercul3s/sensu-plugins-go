package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/portertech/sensu-plugins-go/lib/check"
)

func main() {
	var (
		url          string
		redirect     bool
		timeout      int
		minBytes     int
		requireBytes int
	)

	c := check.New("CheckHTTP")
	c.Option.StringVarP(&url, "url", "u", "http://localhost/", "URL")
	c.Option.BoolVarP(&redirect, "redirect", "r", false, "REDIRECT")
	c.Option.IntVarP(&timeout, "timeout", "t", 15, "TIMEOUT")
	c.Option.IntVarP(&minBytes, "min-bytes", "g", -1, "MIN RESPONSE BYTES")
	c.Option.IntVarP(&requireBytes, "require-bytes", "B", -1, "REQUIRED RESPONSE BYTES")
	c.Init()

	status, bytesRead, err := statusCode(url, timeout, redirect)
	if err != nil {
		c.Error(err)
	}

	switch {
	case bytesRead < minBytes && minBytes > -1:
		c.Critical(fmt.Sprintf("Response was %d bytes instead of minimum of %d bytes", bytesRead, minBytes))
	case bytesRead != requireBytes && requireBytes > -1:
		c.Critical(fmt.Sprintf("Response was %d bytes instead of required %d bytes", bytesRead, requireBytes))
	}

	switch {
	case status >= 400:
		c.Critical(strconv.Itoa(status))
	case status >= 300 && redirect:
		c.Ok(strconv.Itoa(status))
	case status >= 300:
		c.Warning(strconv.Itoa(status))
	default:
		c.Ok(strconv.Itoa(status))
	}
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func statusCode(url string, timeout int, redirect bool) (int, int, error) {
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// ensure client does not follow redirects if redirect flag is set
	if redirect {
		client.CheckRedirect = checkRedirect
	}

	response, err := client.Get(url)
	if err != nil {
		return 0, 0, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, 0, err
	}

	return response.StatusCode, len(body), nil
}
