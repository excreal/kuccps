package functions

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

type ClientConfig struct {
	Timeout          time.Duration
	DialTimeout      time.Duration
	MaxIdleConns     int
	ForceAttemptHTTP2 bool
	MaxConnsPerHost  int
}

var (
	client        *http.Client
	transport     *http.Transport
	dnsResolver   *net.Resolver
	clientOnce    sync.Once
	csrfToken     string
	csrfCookie    string
	csrfMutex     sync.Mutex
	errorRegex    = regexp.MustCompile(`(?i)alert alert-danger`)
)

var (
	ErrTimeout            = errors.New("timed out: check your internet connection")
	ErrCSRFToken          = errors.New("server configuration error: missing security token")
	ErrInvalidCredentials = errors.New("incorrect login details")
	ErrGeneric            = errors.New("server error: try again later")
)

type LoginResult struct {
	Status string
	Error  error
}

func InitializeClient(config ClientConfig) {
	clientOnce.Do(func() {
		dnsResolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				dialer := &net.Dialer{Timeout: 5 * time.Second}
				return dialer.DialContext(ctx, "udp", "8.8.8.8:53")
			},
		}

		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			MaxIdleConns:        config.MaxIdleConns,
			MaxIdleConnsPerHost: config.MaxIdleConns,
			MaxConnsPerHost:     config.MaxConnsPerHost,
			ForceAttemptHTTP2:   config.ForceAttemptHTTP2,
			DialContext: (&net.Dialer{
				Timeout:   config.DialTimeout,
				KeepAlive: 30 * time.Second,
				Resolver:  dnsResolver,
			}).DialContext,
			IdleConnTimeout: 90 * time.Second,
		}

		client = &http.Client{
			Transport: transport,
			Timeout:   config.Timeout,
		}
	})
}

func getCSRFTokenAndCookie() (string, string, error) {
	csrfMutex.Lock()
	defer csrfMutex.Unlock()

	if csrfToken != "" && csrfCookie != "" {
		return csrfToken, csrfCookie, nil
	}

	resp, err := client.Get("https://students.kuccps.net/login/")
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrTimeout, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", ErrGeneric
	}

	token, cookie := extractCSRFTokens(resp, body)
	if token == "" || cookie == "" {
		return "", "", ErrCSRFToken
	}

	csrfToken = token
	csrfCookie = cookie
	return csrfToken, csrfCookie, nil
}

func extractCSRFTokens(resp *http.Response, body []byte) (string, string) {
	re := regexp.MustCompile(`name=['"]csrfmiddlewaretoken['"] value=['"]([^'"]+)['"]`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", ""
	}

	token := matches[1]
	var cookie string
	for _, c := range resp.Cookies() {
		if c.Name == "csrftoken" {
			cookie = c.Value
			break
		}
	}
	return token, cookie
}

func Login(kcseIndexNumber, password, year string) LoginResult {
	token, cookie, err := getCSRFTokenAndCookie()
	if err != nil {
		return LoginResult{Error: err}
	}

	body := fmt.Sprintf(
		"csrfmiddlewaretoken=%s&kcse_index_number=%s&kcse_year=%s&password=%s",
		url.QueryEscape(token),
		url.QueryEscape(kcseIndexNumber),
		url.QueryEscape(year),
		url.QueryEscape(password),
	)

	req, _ := http.NewRequest("POST", "https://students.kuccps.net/login/", strings.NewReader(body))
	req.Header = http.Header{
		"User-Agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"},
		"Accept-Language": {"en-US,en;q=0.9"},
		"Cookie":          {"csrftoken=" + cookie},
		"Content-Type":    {"application/x-www-form-urlencoded"},
		"Referer":         {"https://students.kuccps.net/login/"},
	}

	resp, err := client.Do(req)
	if err != nil {
		return LoginResult{Error: ErrTimeout}
	}
	defer resp.Body.Close()

	bodyContent, _ := ioutil.ReadAll(resp.Body)
	if errorRegex.Match(bodyContent) {
		return LoginResult{Status: "failed", Error: ErrInvalidCredentials}
	}

	return LoginResult{
		Status: determineStatus(resp.StatusCode),
	}
}

func determineStatus(code int) string {
	switch code {
	case http.StatusOK:
		return "valid_session"
	case http.StatusFound:
		return "redirect_received"
	default:
		return "unexpected_response"
	}
}