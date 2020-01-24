package main

import (
	"crypto/tls"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/golang/glog"
)

type tlsKeypairReloader struct {
	certMutex sync.RWMutex
	cert      *tls.Certificate
	certPath  string
	keyPath   string
}

func (keyPair *tlsKeypairReloader) maybeReload() error {
	newCert, err := tls.LoadX509KeyPair(keyPair.certPath, keyPair.keyPath)
	if err != nil {
		return err
	}
	glog.Infoln("cetificate reloaded")
	keyPair.certMutex.Lock()
	defer keyPair.certMutex.Unlock()
	keyPair.cert = &newCert
	return nil
}

func (keyPair *tlsKeypairReloader) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		keyPair.certMutex.RLock()
		defer keyPair.certMutex.RUnlock()
		return keyPair.cert, nil
	}
}

func NewTlsKeypairReloader(certPath, keyPath string) (*tlsKeypairReloader, error) {
	result := &tlsKeypairReloader{
		certPath: certPath,
		keyPath:  keyPath,
	}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	result.cert = &cert

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP)
		for range c {
			if err := result.maybeReload(); err != nil {
				glog.Fatalf("Failed to reload certificate: %v", err)
			}
		}
	}()
	return result, nil
}
