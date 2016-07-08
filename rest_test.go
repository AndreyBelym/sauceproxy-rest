package admin

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

const versionJson = `{
    "Sauce Connect": {
        "download_url": "https://wiki.saucelabs.com/display/DOCS/Setting+Up+Sauce+Connect",
        "linux": {
            "build": 42,
            "download_url": "https://saucelabs.com/downloads/sc-new",
            "sha1": "123456"
        },
        "linux32": {
            "build": 42,
            "download_url": "https://saucelabs.com/downloads/sc-new",
            "sha1": "123456"
        },
        "osx": {
            "build": 42,
            "download_url": "https://saucelabs.com/downloads/sc-new",
            "sha1": "123456"
        },
        "version": "4.3.16",
        "win32": {
            "build": 42,
            "download_url": "https://saucelabs.com/downloads/sc-new",
            "sha1": "123456"
        }
    },
    "Sauce Connect 2": {
        "download_url": "https://docs.saucelabs.com/reference/sauce-connect/",
        "version": "4.3.13-r999"
    }
}`

// Helper function to create a fake http server
func makeServer(f func(w http.ResponseWriter)) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f(w)
		}))
}

func multiResponseServer(responses []string) *httptest.Server {
	var index = 0
	return makeServer(func(w http.ResponseWriter) {
		io.WriteString(w, responses[index])
		if index < len(responses) {
			index += 1;
		}
	})
}

func TestGetLastVersion(t *testing.T) {
	var server = makeServer(func(w http.ResponseWriter) {
		fmt.Fprintln(w, versionJson)
	})
	defer server.Close()

	build, url, err := GetLastVersion(server.URL, &http.Client{})

	if err != nil {
		t.Errorf("%v", err)
	}
	if build != 42 {
		t.Errorf("Bad build number: %d", build)
	}
	if url != "https://saucelabs.com/downloads/sc-new" {
		t.Errorf("Bad URL: %s", url)
	}
}

func TestGetLastVersionBadJSON(t *testing.T) {
	var server = makeServer(func(w http.ResponseWriter) {
		fmt.Fprintln(w, "Not a JSON document...")
	})
	defer server.Close()

	_, _, err := GetLastVersion(server.URL, &http.Client{})

	if err == nil {
		t.Error("GetLastVersion == nil")
	}

	if !strings.HasPrefix(err.Error(), "couldn't decode JSON document: ") {
		t.Errorf("Invalid error: %s", err.Error())
	}
}

func TestGetLastVersion404(t *testing.T) {
	var server = makeServer(func(w http.ResponseWriter) {
		http.Error(w, "nothing to see here", 404)
	})
	defer server.Close()

	_, _, err := GetLastVersion(server.URL, &http.Client{})

	if err == nil {
		t.Error("GetLastVersion == nil")
	}

	if !strings.HasPrefix(err.Error(), "couldn't find ") {
		t.Errorf("Invalid error: %s", err.Error())
	}
}

func TestGetLastVersionNoServer(t *testing.T) {
	var server = makeServer(func(w http.ResponseWriter) {})
	// We close the server right-away so it doesn't response to requests, but we
	// still keep it around so our client has a 'bad' URL to connect to.
	server.Close()

	_, _, err := GetLastVersion(server.URL, &http.Client{})

	if err == nil {
		t.Error("GetLastVersion == nil")
	}

	if !strings.HasPrefix(err.Error(), "couldn't connect to ") {
		t.Errorf("Invalid error: %s", err.Error())
	}
}

const tunnelsJSON = `[
  {
    "status": "running",
    "direct_domains": null,
    "vm_version": null,
    "last_connected": 1467691618,
    "shutdown_time": null,
    "ssh_port": 443,
    "launch_time": 1467690963,
    "user_shutdown": null,
    "use_caching_proxy": null,
    "creation_time": 1467690959,
    "domain_names": [
      "sauce-connect.proxy"
    ],
    "shared_tunnel": false,
    "tunnel_identifier": null,
    "host": "maki81134.miso.saucelabs.com",
    "no_proxy_caching": false,
    "owner": "henryprecheur",
    "use_kgp": true,
    "no_ssl_bump_domains": null,
    "id": "fakeid",
    "metadata": {
      "hostname": "debian-desktop",
      "git_version": "39e807b",
      "platform": "Linux 4.6.0-1-amd64 #1 SMP Debian 4.6.2-2 (2016-06-25) x86_64",
      "command": "./sc -u henryprecheur -k ****",
      "build": "2396",
      "release": "4.3.16",
      "nofile_limit": 1024
    }
  }
]`

func TestClientMatch(t *testing.T) {
	var server = makeServer(func(w http.ResponseWriter) {
		fmt.Fprintln(w, tunnelsJSON)
	})
	defer server.Close()

	var client = Client{
		BaseURL:  server.URL,
		Username: "username",
		Password: "password",
	}

	var matches, err = client.Match("fakeid", []string{"sauce-connect.proxy"})

	if err != nil {
		t.Errorf("client.Match errored %+v\n", err)
	}

	if !reflect.DeepEqual(matches, []string{"fakeid"}) {
		t.Errorf("client.Match returned %+v\n", matches)
	}
}

func TestClientShutdown(t *testing.T) {
	var server = makeServer(func(w http.ResponseWriter) {
		fmt.Fprintln(w, "")
	})
	defer server.Close()

	var client = Client{
		BaseURL:  server.URL,
		Username: "username",
		Password: "password",
	}

	err := client.Shutdown("fakeid")
	if err != nil {
		t.Errorf("client.Shutdown errored %+v\n", err)
	}
}

func TestClientShutdown404(t *testing.T) {
	var server = makeServer(func(w http.ResponseWriter) {
		http.Error(w, "nothing to see here", 404)
	})
	defer server.Close()

	var client = Client{
		BaseURL:  server.URL,
		Username: "username",
		Password: "password",
	}

	err := client.Shutdown("fakeid")
	if !strings.HasPrefix(err.Error(), "couldn't find ") {
		t.Errorf("Invalid error: %s", err.Error())
	}
}

const createJSON = `{
  "status": "new",
  "direct_domains": null,
  "vm_version": null,
  "last_connected": null,
  "shutdown_time": null,
  "ssh_port": 443,
  "launch_time": null,
  "user_shutdown": null,
  "use_caching_proxy": null,
  "creation_time": 1467839998,
  "domain_names": [
    "sauce-connect.proxy"
  ],
  "shared_tunnel": false,
  "tunnel_identifier": null,
  "host": null,
  "no_proxy_caching": false,
  "owner": "henryprecheur",
  "use_kgp": true,
  "no_ssl_bump_domains": null,
  "id": "49958ce5ec9f49c796542e0c691455a6",
  "metadata": {
    "hostname": "Henry's computer",
    "git_version": "4a804fd",
    "platform": "plan9",
    "command": "./sc",
    "build": "Strong",
    "release": "1.2.3",
    "no_file_limit": 12345
  }
}`

const statusRunningJSON = `{"status": "running", "user_shutdown": null}`

func TestClientCreate(t *testing.T) {
	var server = multiResponseServer([]string{createJSON, statusRunningJSON})
	defer server.Close()

	var client = Client{
		BaseURL:  server.URL,
		Username: "username",
		Password: "password",
	}
	var request = Request{
		DomainNames: []string{"sauce-connect.proxy"},
	}
	_, err := client.createWithTimeouts(&request, time.Second, time.Second)
	if err != nil {
		t.Errorf("client.createWithTimeouts errored %+v\n", err)
	}
}

func TestClientCreateError(t *testing.T) {
	var server = multiResponseServer([]string{"ERROR!"})
	defer server.Close()

	var client = Client{
		BaseURL:  server.URL,
		Username: "username",
		Password: "password",
	}
	var request = Request{
		DomainNames: []string{"sauce-connect.proxy"},
	}
	_, err := client.createWithTimeouts(&request, time.Second, 0)
	if err == nil {
		t.Errorf("client.createWithTimeouts didn't error")
	}

	if !(
		strings.HasPrefix(err.Error(), "couldn't decode JSON document: ")) {
		t.Errorf("Invalid error: %s", err.Error())
	}
}

func TestClientCreateWaitError(t *testing.T) {
	var server = multiResponseServer(
		[]string{createJSON, `{"status": "shutdown", "user_shutdown": null}`})
	defer server.Close()

	var client = Client{
		BaseURL:  server.URL,
		Username: "username",
		Password: "password",
	}
	var request = Request{
		DomainNames: []string{"sauce-connect.proxy"},
	}
	_, err := client.createWithTimeouts(&request, time.Second, 0)
	if err == nil {
		t.Errorf("client.createWithTimeouts didn't error")
	}

	if !(
		strings.HasPrefix(err.Error(), "Tunnel ") &&
		strings.HasSuffix(err.Error(), " didn't come up after 0")) {
		t.Errorf("Invalid error: %s", err.Error())
	}
}
