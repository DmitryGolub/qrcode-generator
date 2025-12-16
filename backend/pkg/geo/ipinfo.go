package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)


type ipinfoLiteClient struct {
	http  *http.Client
	token string
	cache *ttlCache[string, ipinfoLiteResp]
	ttl   time.Duration
}

type ipinfoLiteResp struct {
	Country string `json:"country"`
}

type ttlCache[K comparable, V any] struct {
	items map[K]struct {
		v   V
		exp time.Time
	}
}

func newTTLCache[K comparable, V any]() *ttlCache[K, V] {
	return &ttlCache[K, V]{items: make(map[K]struct {
		v   V
		exp time.Time
	})}
}

func (c *ttlCache[K, V]) Get(k K) (V, bool) {
	it, ok := c.items[k]
	if !ok || time.Now().After(it.exp) {
		var zero V
		return zero, false
	}
	return it.v, true
}

func (c *ttlCache[K, V]) Set(k K, v V, ttl time.Duration) {
	c.items[k] = struct {
		v   V
		exp time.Time
	}{v: v, exp: time.Now().Add(ttl)}
}

func newIPInfoLite(token string, timeout time.Duration) *ipinfoLiteClient {
	return &ipinfoLiteClient{
		http:  &http.Client{Timeout: timeout},
		token: token,
		cache: newTTLCache[string, ipinfoLiteResp](),
		ttl:   10 * time.Minute,
	}
}

func (c *ipinfoLiteClient) Resolve(ip string) (string, string, bool) {
	if net.ParseIP(ip) == nil {
		return "", "", false
	}
	key := ipSubnet24(ip)

	if cached, ok := c.cache.Get(key); ok {
		if cached.Country != "" {
			return cached.Country, "", true
		}
		return "", "", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.http.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("https://ipinfo.io/%s/json?token=%s", ip, c.token), nil)
	if err != nil {
		return "", "", false
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.cache.Set(key, ipinfoLiteResp{}, 1*time.Minute)
		return "", "", false
	}

	var data ipinfoLiteResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", false
	}

	c.cache.Set(key, data, c.ttl)
	if data.Country == "" {
		return "", "", false
	}
	return data.Country, "", true
}

func ipSubnet24(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}
	v4 := parsed.To4()
	if v4 == nil {
		return parsed.String()[:12]
	}
	return fmt.Sprintf("%d.%d.%d.0/24", v4[0], v4[1], v4[2])
}
