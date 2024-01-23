package cache

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cygkam/dcache/pkg/cachepool"
	gubernator "github.com/mailgun/gubernator/v2"
	"github.com/sirupsen/logrus"
)

type CacheConfig struct {
	DistributionEnabled bool
	Namespace           string
	PodIP               string
	Port                string
	Selector            string
}

type Cache struct {
	CachePool *cachepool.CachePool
	hostname  string
	k8sPool   *gubernator.K8sPool
}

func New(cfg *CacheConfig, f cachepool.Fetcher) (*Cache, error) {
	hostname := fmt.Sprintf("%v:%v", cfg.PodIP, cfg.Port)
	cachePoolCfg := &cachepool.CachePoolCfg{
		Ttl:     time.Second * 30,
		Fetcher: f,
		Port:    cfg.Port,
	}
	cp := cachepool.NewCachePool(cachePoolCfg)
	cp.SetPeers(hostname)

	cache := &Cache{
		CachePool: cp,
		hostname:  hostname,
	}

	if cfg.DistributionEnabled {
		if err := cache.watchPods(cfg.Namespace, cfg.Selector, cfg.PodIP, cfg.Port); err != nil {
			return nil, err
		}
	}

	return cache, nil
}

func (c *Cache) StartHTTPPoolServer() *http.Server {
	server := &http.Server{
		Addr:    c.hostname,
		Handler: c.CachePool,
	}

	logrus.Info("Starting HTTP Cache Server")
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Cannot start HTTP Cache Server, error: %v", err)
		}
	}()

	return server
}

func (c *Cache) watchPods(namespace string, selector string, podIP string, port string) error {
	k8sPoolConfig := gubernator.K8sPoolConfig{
		Mechanism: gubernator.WatchPods,
		OnUpdate: func(infos []gubernator.PeerInfo) {
			updatePeersPool(c.CachePool, infos)
		},
		Namespace: namespace,
		Selector:  selector,
		PodIP:     podIP,
		PodPort:   port,
	}

	var err error
	c.k8sPool, err = gubernator.NewK8sPool(k8sPoolConfig)
	if err != nil {
		return err
	}

	return nil
}

func updatePeersPool(cp *cachepool.CachePool, infos []gubernator.PeerInfo) {
	var peers []string

	for _, info := range infos {
		var addr string

		if info.HTTPAddress != "" {
			addr = info.HTTPAddress
		} else if info.GRPCAddress != "" {
			addr = info.GRPCAddress

			if !strings.HasPrefix(addr, "http") {
				addr = "http://" + strings.Trim(addr, "/")
			}
		}

		if addr == "" {
			logrus.Warnf("Missing address for peer info: %v", info)
			break
		}

		u, err := url.Parse(addr)
		if err != nil {
			logrus.Warnf("Error parsing peer address: %v, %v", addr, err)
		} else {
			u.Scheme = "http"

			peer := u.String()
			logrus.Debugf("Peer found: %v, current instance: %v", peer, info.IsOwner)
			peers = append(peers, peer)
		}
	}
	logrus.Infof("Peers pool: %v, count: %v", peers, len(peers))
	cp.SetPeers(peers...)
}
