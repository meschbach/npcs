// Package qcpki provides quality control mechanisms for public key infrastructure
package qcpki

import (
	"crypto/tls"
	"github.com/madflojo/testcerts"
)

type RootCA struct {
	root *testcerts.CertificateAuthority
}

func NewRootCA() *RootCA {
	return &RootCA{
		root: testcerts.NewCA(),
	}
}

func (r *RootCA) NewServiceCertificate(domainName string) (*ServiceCertificate, error) {
	pair, err := r.root.NewKeyPair(domainName)
	if err != nil {
		return nil, err
	}
	cfg, err := pair.ConfigureTLSConfig(r.root.GenerateTLSConfig())
	if err != nil {
		return nil, err
	}
	return &ServiceCertificate{cfg: cfg, pair: pair}, nil
}

type ServiceCertificate struct {
	pair *testcerts.KeyPair
	cfg  *tls.Config
}

func (s *ServiceCertificate) TLSConfig() *tls.Config {
	return s.cfg
}
