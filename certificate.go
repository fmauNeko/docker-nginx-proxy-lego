package main

import (
	"strings"

	log "github.com/Sirupsen/logrus"
)

// Certificate represents a certificate to be generated
type Certificate struct {
	container *Container
	account   *Account
}

func (cert *Certificate) generateCertificate() {
	log.WithFields(log.Fields{"Hosts": strings.Join(cert.container.hosts, " ")}).Info("Generating new certificate")

	// We're done here !
	checkWg.Done()
}
