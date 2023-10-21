package cdn

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/bnet/bpsv"
	"github.com/Dakota628/d4parse/pkg/bnet/ribbit2"
	"github.com/avast/retry-go"
	"leb.io/hashland/jenkins"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Endpoint ...
type Endpoint struct {
	URL      *url.URL
	MaxHosts int
	Fallback bool
}

func (e *Endpoint) BuildURL(path string, pathType string, hexHash string) string {
	return e.URL.JoinPath(path, pathType, hexHash[0:2], hexHash[2:4], hexHash).String()
}

func ParseEndpoint(s string) (e *Endpoint, err error) {
	e = &Endpoint{}

	e.URL, err = url.Parse(s)
	if err != nil {
		return nil, err
	}

	query := e.URL.Query()

	if query.Has("maxhosts") {
		e.MaxHosts, err = strconv.Atoi(query.Get("maxhosts"))
		if err != nil {
			return nil, err
		}
	}

	e.Fallback = query.Get("fallback") == "1"

	return e, nil
}

// Servers ...
type Servers struct {
	Primaries []*Endpoint
	Fallbacks []*Endpoint
}

func (s *Servers) selectEndpoint(endpoints []*Endpoint) *Endpoint {
	switch len(endpoints) {
	case 0:
		return nil
	case 1:
		return endpoints[0]
	default:
		return endpoints[rand.Intn(len(endpoints))]
	}
}

func (s *Servers) SelectPrimary() *Endpoint {
	return s.selectEndpoint(s.Primaries)
}

func (s *Servers) SelectFallback() *Endpoint {
	return s.selectEndpoint(s.Fallbacks)
}

func ParseServers(cdn bpsv.Row) (*Servers, error) {
	servers := &Servers{}

	for _, server := range strings.Split(cdn["Servers"], " ") {
		e, err := ParseEndpoint(server)
		if err != nil {
			return nil, err
		}

		if e.Fallback {
			servers.Fallbacks = append(servers.Fallbacks, e)
		} else {
			servers.Primaries = append(servers.Primaries, e)
		}
	}

	return servers, nil
}

func FetchCDNInfo(ribbit *ribbit2.Client, product string, region string) (bpsv.Row, error) {
	command := bytes.Buffer{}
	command.Grow(12 + len(product) + 5)
	command.WriteString("v2/products/")
	command.WriteString(product)
	command.WriteString("/cdns")

	resp, err := ribbit.Do(ribbit2.Request{
		Command: command.Bytes(),
	})

	if err != nil {
		return bpsv.Row{}, err
	}

	doc, err := resp.BPSV()
	if err != nil {
		return bpsv.Row{}, err
	}

	for _, row := range doc.Rows {
		if row["Name"] == region {
			return row, nil
		}
	}

	return bpsv.Row{}, errors.New("no CDN for region")
}

func FetchVersionInfo(ribbit *ribbit2.Client, product string, region string) (bpsv.Row, error) {
	command := bytes.Buffer{}
	command.Grow(12 + len(product) + 9)
	command.WriteString("v2/products/")
	command.WriteString(product)
	command.WriteString("/versions")

	resp, err := ribbit.Do(ribbit2.Request{
		Command: command.Bytes(),
	})

	if err != nil {
		return bpsv.Row{}, err
	}

	doc, err := resp.BPSV()
	if err != nil {
		return bpsv.Row{}, err
	}

	for _, row := range doc.Rows {
		if row["Region"] == region {
			return row, nil
		}
	}

	return bpsv.Row{}, errors.New("no version for region")
}

// CDN ...
type CDN struct {
	httpClient   *http.Client
	ribbitClient *ribbit2.Client

	servers    *Servers
	configPath string
	path       string

	version bpsv.Row
}

func NewCDN(ribbit *ribbit2.Client, product string, region string) (*CDN, error) {
	// Get CDN info via Ribbit
	cdn, err := FetchCDNInfo(ribbit, product, region)
	if err != nil {
		return nil, err
	}

	// Parse servers from CDN info into servers objects
	servers, err := ParseServers(cdn)
	if err != nil {
		return nil, err
	}

	// Get version info via Ribbit
	version, err := FetchVersionInfo(ribbit, product, region)
	if err != nil {
		return nil, err
	}

	return &CDN{
		httpClient:   &http.Client{},
		ribbitClient: ribbit,

		servers:    servers,
		configPath: cdn["ConfigPath"],
		path:       cdn["Path"],

		version: version,
	}, nil
}

func (c *CDN) Version() string {
	return c.version["VersionsName"]
}

func (c *CDN) GetBuildConfig() (*http.Response, error) {
	return c.Do("config", c.version["BuildConfig"])
}

func (c *CDN) GetCDNConfig() (*http.Response, error) {
	return c.Do("config", c.version["CDNConfig"])
}

func (c *CDN) GetProductConfig() (*http.Response, error) {
	return c.Do("config", c.version["ProductConfig"])
}

func (c *CDN) Do(pathType string, hexHash string) (*http.Response, error) { // TODO: may want ctx
	var fallback bool
	var resp *http.Response

	if err := retry.Do(
		func() (err error) {
			// Select an endpoint for this retry
			var endpoint *Endpoint
			if fallback {
				endpoint = c.servers.SelectFallback()
			} else {
				endpoint = c.servers.SelectPrimary()
				fallback = true
			}

			// Try to send the request
			reqUrl := endpoint.BuildURL(c.path, pathType, hexHash)
			resp, err = c.httpClient.Get(reqUrl)
			if err != nil {
				return
			}

			if resp.StatusCode >= 300 {
				return fmt.Errorf("not ok (%d)", resp.StatusCode)
			}

			return
		},
		retry.Attempts(10), // TODO: better retry policy?
	); err != nil {
		return nil, err
	}

	return resp, nil
}

func NormalizeFileName(fileName string) string {
	return strings.ReplaceAll(strings.ToUpper(fileName), "/", "\\")
}

func NameHash(fileName string) uint32 {
	fileName = NormalizeFileName(fileName)
	rpc, _ := jenkins.HashString(fileName, 0, 0)
	return rpc
}
