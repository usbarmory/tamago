// https://github.com/inversepath/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build tamago,arm

package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func generateTLSCerts(address net.IP) ([]byte, []byte, error) {
	TLSCert := new(bytes.Buffer)
	TLSKey := new(bytes.Buffer)

	serial, _ := rand.Int(rand.Reader, big.NewInt(1<<63-1))

	log.Printf("imx6_tls: generating TLS keypair IP: %s, Serial: %X", IP, serial)

	validFrom, _ := time.Parse(time.RFC3339, "1981-01-07T00:00:00Z")
	validUntil, _ := time.Parse(time.RFC3339, "2022-01-07T00:00:00Z")

	certTemplate := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization:       []string{"F-Secure Foundry"},
			OrganizationalUnit: []string{"TamaGo test certificates"},
			CommonName:         IP,
		},
		IPAddresses:        []net.IP{address},
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		PublicKeyAlgorithm: x509.ECDSA,
		NotBefore:          validFrom,
		NotAfter:           validUntil,
		SubjectKeyId:       []byte{1, 2, 3, 4, 5},
		KeyUsage:           x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	caTemplate := certTemplate
	caTemplate.SerialNumber = serial
	caTemplate.SubjectKeyId = []byte{1, 2, 3, 4, 6}
	caTemplate.BasicConstraintsValid = true
	caTemplate.IsCA = true
	caTemplate.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign
	caTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}

	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pub := &priv.PublicKey
	cert, err := x509.CreateCertificate(rand.Reader, &certTemplate, &caTemplate, pub, priv)

	if err != nil {
		return nil, nil, err
	}

	pem.Encode(TLSCert, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
	ecb, _ := x509.MarshalECPrivateKey(priv)
	pem.Encode(TLSKey, &pem.Block{Type: "EC PRIVATE KEY", Bytes: ecb})

	h := sha256.New()
	h.Write(cert)

	log.Printf("imx6_tls: SHA-256 fingerprint: % X", h.Sum(nil))

	return TLSCert.Bytes(), TLSKey.Bytes(), nil
}

func startWebServer(s *stack.Stack, addr tcpip.Address, port uint16, nic tcpip.NICID, https bool) {
	var err error

	fullAddr := tcpip.FullAddress{Addr: addr, Port: port, NIC: nic}
	listener, err := gonet.NewListener(s, fullAddr, ipv4.ProtocolNumber)

	if err != nil {
		log.Fatal("listener error: ", err)
	}

	srv := &http.Server{
		Addr: addr.String() + ":" + string(port),
	}

	if https {
		TLSCert, TLSKey, err := generateTLSCerts(net.ParseIP(addr.String()))

		if err != nil {
			log.Fatal("TLS cert|key error: ", err)
		}

		log.Printf("%s", TLSCert)
		log.Printf("%s", TLSKey)

		certificate, err := tls.X509KeyPair(TLSCert, TLSKey)

		if err != nil {
			log.Fatal("X509KeyPair error: ", err)
		}

		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}
	}

	log.Printf("imx6_web: starting web server at %s:%d", addr.String(), port)

	if https {
		err = srv.ServeTLS(listener, "", "")
	} else {
		err = srv.Serve(listener)
	}

	log.Fatal("server returned unexpectedly ", err)

	return
}
