package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/xenolf/lego/acme"
)

// Configuration type from Env.
type Configuration struct {
	dEndpoint string
	keyTypes  []string
	leServer  string
	path      string
}

// NewConfiguration creates a new configuration from CLI data.
func NewConfiguration() *Configuration {
	var c Configuration

	if dEndpoint := os.Getenv("DOCKER_ENDPOINT"); len(dEndpoint) > 0 {
		c.dEndpoint = dEndpoint
	} else {
		c.dEndpoint = "unix:///var/run/docker.sock"
	}

	if keyTypes := os.Getenv("LETSENCRYPT_KEYTYPES"); len(keyTypes) > 0 {
		c.keyTypes = strings.Split(keyTypes, "+")
	} else {
		c.keyTypes = []string{"EC384", "RSA4096"}
	}

	if leServer := os.Getenv("LETSENCRYPT_SERVER"); len(leServer) > 0 {
		c.leServer = leServer
	} else {
		c.leServer = "https://acme-v01.api.letsencrypt.org/directory"
	}

	if dataPath := os.Getenv("LETSENCRYPT_PATH"); len(dataPath) > 0 {
		c.path = dataPath
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			log.Panic("Failed to get working directory")
		}
		c.path = path.Join(cwd, ".lego")
	}

	log.WithField("conf", c).Info("Config")

	return &c
}

// KeyTypes the types from which private keys should be generated
func (c *Configuration) KeyTypes() ([]acme.KeyType, error) {
	keyTypes := make([]acme.KeyType, len(c.keyTypes))

	for i, kt := range c.keyTypes {
		keyType, err := keyType(kt)
		if err != nil {
			return nil, err
		}
		keyTypes[i] = keyType
	}

	return keyTypes, nil
}

// KeyType the type from which private keys should be generated
func keyType(kt string) (acme.KeyType, error) {
	switch strings.ToUpper(kt) {
	case "RSA2048":
		return acme.RSA2048, nil
	case "RSA4096":
		return acme.RSA4096, nil
	case "RSA8192":
		return acme.RSA8192, nil
	case "EC256":
		return acme.EC256, nil
	case "EC384":
		return acme.EC384, nil
	}

	return "", fmt.Errorf("Unsupported KeyType: %s", kt)
}

// LEServerPath returns the OS dependent path to the data for a specific CA
func (c *Configuration) LEServerPath() string {
	srv, err := url.Parse(c.leServer)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "leServer": c.leServer}).Panic("Failed to parse leServer URL")
	}
	srvStr := strings.Replace(srv.Host, ":", "_", -1)
	return strings.Replace(srvStr, "/", string(os.PathSeparator), -1)
}

// CertPath gets the path for certificates.
func (c *Configuration) CertPath() string {
	return path.Join(c.path, "certificates")
}

// AccountsPath returns the OS dependent path to the
// local accounts for a specific CA
func (c *Configuration) AccountsPath() string {
	return path.Join(c.path, "accounts", c.LEServerPath())
}

// AccountPath returns the OS dependent path to a particular account
func (c *Configuration) AccountPath(acc string) string {
	return path.Join(c.AccountsPath(), acc)
}

// AccountKeysPath returns the OS dependent path to the keys of a particular account
func (c *Configuration) AccountKeysPath(acc string) string {
	return path.Join(c.AccountPath(acc), "keys")
}
