/*
Copyright 2014 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func parseURLOrDie(inURL string) *url.URL {
	parsed, err := url.Parse(inURL)
	if err != nil {
		panic(err)
	}
	return parsed
}

// fmtHTML parses and re-emits 'in', effectively canonicalizing it.
func fmtHTML(in string) string {
	doc, err := html.Parse(strings.NewReader(in))
	if err != nil {
		panic(err)
	}
	out := &bytes.Buffer{}
	if err := html.Render(out, doc); err != nil {
		panic(err)
	}
	return string(out.Bytes())
}

func TestProxyTransport(t *testing.T) {
	testTransport := &Transport{
		Scheme:      "http",
		Host:        "foo.com",
		PathPrepend: "/proxy/minion/minion1:10250",
	}
	testTransport2 := &Transport{
		Scheme:      "https",
		Host:        "foo.com",
		PathPrepend: "/proxy/minion/minion1:8080",
	}
	type Item struct {
		input        string
		sourceURL    string
		transport    *Transport
		output       string
		contentType  string
		forwardedURI string
		redirect     string
		redirectWant string
	}

	table := map[string]Item{
		"normal": {
			input:        `<pre><a href="kubelet.log">kubelet.log</a><a href="/google.log">google.log</a></pre>`,
			sourceURL:    "http://myminion.com/logs/log.log",
			transport:    testTransport,
			output:       `<pre><a href="http://foo.com/proxy/minion/minion1:10250/logs/kubelet.log">kubelet.log</a><a href="http://foo.com/proxy/minion/minion1:10250/google.log">google.log</a></pre>`,
			contentType:  "text/html",
			forwardedURI: "/proxy/minion/minion1:10250/logs/log.log",
		},
		"trailing slash": {
			input:        `<pre><a href="kubelet.log">kubelet.log</a><a href="/google.log/">google.log</a></pre>`,
			sourceURL:    "http://myminion.com/logs/log.log",
			transport:    testTransport,
			output:       `<pre><a href="http://foo.com/proxy/minion/minion1:10250/logs/kubelet.log">kubelet.log</a><a href="http://foo.com/proxy/minion/minion1:10250/google.log/">google.log</a></pre>`,
			contentType:  "text/html",
			forwardedURI: "/proxy/minion/minion1:10250/logs/log.log",
		},
		"content-type charset": {
			input:        `<pre><a href="kubelet.log">kubelet.log</a><a href="/google.log">google.log</a></pre>`,
			sourceURL:    "http://myminion.com/logs/log.log",
			transport:    testTransport,
			output:       `<pre><a href="http://foo.com/proxy/minion/minion1:10250/logs/kubelet.log">kubelet.log</a><a href="http://foo.com/proxy/minion/minion1:10250/google.log">google.log</a></pre>`,
			contentType:  "text/html; charset=utf-8",
			forwardedURI: "/proxy/minion/minion1:10250/logs/log.log",
		},
		"content-type passthrough": {
			input:        `<pre><a href="kubelet.log">kubelet.log</a><a href="/google.log">google.log</a></pre>`,
			sourceURL:    "http://myminion.com/logs/log.log",
			transport:    testTransport,
			output:       `<pre><a href="kubelet.log">kubelet.log</a><a href="/google.log">google.log</a></pre>`,
			contentType:  "text/plain",
			forwardedURI: "/proxy/minion/minion1:10250/logs/log.log",
		},
		"subdir": {
			input:        `<a href="kubelet.log">kubelet.log</a><a href="/google.log">google.log</a>`,
			sourceURL:    "http://myminion.com/whatever/apt/somelog.log",
			transport:    testTransport2,
			output:       `<a href="https://foo.com/proxy/minion/minion1:8080/whatever/apt/kubelet.log">kubelet.log</a><a href="https://foo.com/proxy/minion/minion1:8080/google.log">google.log</a>`,
			contentType:  "text/html",
			forwardedURI: "/proxy/minion/minion1:8080/whatever/apt/somelog.log",
		},
		"image": {
			input:        `<pre><img src="kubernetes.jpg"/></pre>`,
			sourceURL:    "http://myminion.com/",
			transport:    testTransport,
			output:       `<pre><img src="http://foo.com/proxy/minion/minion1:10250/kubernetes.jpg"/></pre>`,
			contentType:  "text/html",
			forwardedURI: "/proxy/minion/minion1:10250/",
		},
		"abs": {
			input:        `<script src="http://google.com/kubernetes.js"/>`,
			sourceURL:    "http://myminion.com/any/path/",
			transport:    testTransport,
			output:       `<script src="http://google.com/kubernetes.js"/>`,
			contentType:  "text/html",
			forwardedURI: "/proxy/minion/minion1:10250/any/path/",
		},
		"abs but same host": {
			input:        `<script src="http://myminion.com/kubernetes.js"/>`,
			sourceURL:    "http://myminion.com/any/path/",
			transport:    testTransport,
			output:       `<script src="http://foo.com/proxy/minion/minion1:10250/kubernetes.js"/>`,
			contentType:  "text/html",
			forwardedURI: "/proxy/minion/minion1:10250/any/path/",
		},
		"redirect rel": {
			sourceURL:    "http://myminion.com/redirect",
			transport:    testTransport,
			redirect:     "/redirected/target/",
			redirectWant: "http://foo.com/proxy/minion/minion1:10250/redirected/target/",
			forwardedURI: "/proxy/minion/minion1:10250/redirect",
		},
		"redirect abs same host": {
			sourceURL:    "http://myminion.com/redirect",
			transport:    testTransport,
			redirect:     "http://myminion.com/redirected/target/",
			redirectWant: "http://foo.com/proxy/minion/minion1:10250/redirected/target/",
			forwardedURI: "/proxy/minion/minion1:10250/redirect",
		},
		"redirect abs other host": {
			sourceURL:    "http://myminion.com/redirect",
			transport:    testTransport,
			redirect:     "http://example.com/redirected/target/",
			redirectWant: "http://example.com/redirected/target/",
			forwardedURI: "/proxy/minion/minion1:10250/redirect",
		},
	}

	testItem := func(name string, item *Item) {
		// Canonicalize the html so we can diff.
		item.input = fmtHTML(item.input)
		item.output = fmtHTML(item.output)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check request headers.
			if got, want := r.Header.Get("X-Forwarded-Uri"), item.forwardedURI; got != want {
				t.Errorf("%v: X-Forwarded-Uri = %q, want %q", name, got, want)
			}
			if got, want := r.Header.Get("X-Forwarded-Host"), item.transport.Host; got != want {
				t.Errorf("%v: X-Forwarded-Host = %q, want %q", name, got, want)
			}
			if got, want := r.Header.Get("X-Forwarded-Proto"), item.transport.Scheme; got != want {
				t.Errorf("%v: X-Forwarded-Proto = %q, want %q", name, got, want)
			}

			// Send response.
			if item.redirect != "" {
				http.Redirect(w, r, item.redirect, http.StatusMovedPermanently)
				return
			}
			w.Header().Set("Content-Type", item.contentType)
			fmt.Fprint(w, item.input)
		}))
		defer server.Close()

		// Replace source URL with our test server address.
		sourceURL := parseURLOrDie(item.sourceURL)
		serverURL := parseURLOrDie(server.URL)
		item.input = strings.Replace(item.input, sourceURL.Host, serverURL.Host, -1)
		item.redirect = strings.Replace(item.redirect, sourceURL.Host, serverURL.Host, -1)
		sourceURL.Host = serverURL.Host

		req, err := http.NewRequest("GET", sourceURL.String(), nil)
		if err != nil {
			t.Errorf("%v: Unexpected error: %v", name, err)
			return
		}
		resp, err := item.transport.RoundTrip(req)
		if err != nil {
			t.Errorf("%v: Unexpected error: %v", name, err)
			return
		}
		if item.redirect != "" {
			// Check that redirect URLs get rewritten properly.
			if got, want := resp.Header.Get("Location"), item.redirectWant; got != want {
				t.Errorf("%v: Location header = %q, want %q", name, got, want)
			}
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%v: Unexpected error: %v", name, err)
			return
		}
		if e, a := item.output, string(body); e != a {
			t.Errorf("%v: expected %v, but got %v", name, e, a)
		}
	}

	for name, item := range table {
		testItem(name, &item)
	}
}
