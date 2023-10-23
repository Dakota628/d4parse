package cdn

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/bin"
	"github.com/Dakota628/d4parse/pkg/bnet/bpsv"
	"github.com/Dakota628/d4parse/pkg/bnet/btle"
	"github.com/Dakota628/d4parse/pkg/bnet/ribbit2"
	"github.com/avast/retry-go"
	"io"
	"leb.io/hashland/jenkins"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrInvalidConfigFormat      = errors.New("invalid config format")
	ErrInvalidConfigValueFormat = errors.New("invalid config value format")
)

// Endpoint ...
type Endpoint struct {
	URL      *url.URL
	MaxHosts int
	Fallback bool
}

func (e *Endpoint) BuildURL(path string, pathType string, hexHash string, suffix ...string) string {
	elem := append([]string{path, pathType, hexHash[0:2], hexHash[2:4], hexHash}, suffix...)
	return e.URL.JoinPath(elem...).String()
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

func (c *CDN) GetBuildConfig() (map[string]string, error) {
	return c.GetConfig("BuildConfig")
}

func (c *CDN) GetCDNConfig() (map[string]string, error) {
	return c.GetConfig("CDNConfig")
}

func (c *CDN) GetProductConfig() (map[string]string, error) {
	return c.GetConfig("ProductConfig")
}

func (c *CDN) GetConfig(config string) (map[string]string, error) {
	resp, err := c.Do("config", c.version[config])
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return ParseConfig(body)
}

func (c *CDN) GetEncodingTable() (*btle.EncodingTable, error) {
	// Get encoding hashes from fetched build config
	// TODO: maybe build config should be a arg here so user can cache?
	buildConfig, err := c.GetBuildConfig()
	if err != nil {
		return nil, err
	}

	encodingHashes := strings.Fields(buildConfig["encoding"])
	if len(encodingHashes) != 2 {
		return nil, ErrInvalidConfigValueFormat
	}

	// Fetch encoding table bytes
	resp, err := c.Do("data", encodingHashes[1])
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body) // TODO: parse as BTLE encoding table
	if err != nil {
		return nil, err
	}

	// Convert the body to a binary reader and parse encoding table
	r := bin.NewBinaryReader(bytes.NewReader(body))

	var t btle.EncodingTable
	if err := t.UnmarshalBinary(r); err != nil {
		return nil, err
	}

	return &t, nil
}

func (c *CDN) Do(pathType string, hexHash string, suffix ...string) (*http.Response, error) { // TODO: may want ctx
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
			reqUrl := endpoint.BuildURL(c.path, pathType, hexHash, suffix...)
			resp, err = c.httpClient.Get(reqUrl)
			if err != nil {
				return
			}

			if resp.StatusCode >= 300 {
				return fmt.Errorf("not ok (%d)", resp.StatusCode)
			}

			return
		},
		retry.Attempts(1), // TODO: better retry policy?
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

func ParseConfig(bs []byte) (map[string]string, error) {
	lines := bytes.Split(bs, []byte{'\n'})

	if len(lines) >= 2 {
		if bytes.HasPrefix(lines[0], []byte{'#'}) {
			lines = lines[1:]
		}

		if len(lines[0]) == 0 {
			lines = lines[1:]
		}
	}

	parsed := make(map[string]string, len(lines))

	for _, line := range lines {
		if bytes.HasPrefix(line, []byte{'#'}) {
			continue
		}

		if len(line) == 0 {
			continue
		}

		parts := bytes.SplitN(line, []byte{' ', '=', ' '}, 2)
		if len(parts) != 2 {
			return nil, ErrInvalidConfigFormat
		}

		parsed[string(parts[0])] = string(parts[1])
	}

	return parsed, nil
}
