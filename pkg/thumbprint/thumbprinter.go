/*
 * Copyright 2022 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package thumbprint

import (
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"
)

func RetrieveSHA1(host string, port int) (string, error) {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), conf)
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close() }()

	// Get the ConnectionState struct as that's the one which gives us x509.Certificate struct
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) < 1 {
		return "", fmt.Errorf("expected at least one certificate returned from %s:%d, but got none", host, port)
	}

	return generateSHA1(certs[0]), nil
}

func generateSHA1(cert *x509.Certificate) string {
	sum := sha1.Sum(cert.Raw)
	hex := make([]string, len(sum))
	for i, b := range sum {
		hex[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(hex, ":")
}
