// nolint: gocyclo
package main

import (
	"os"
	"reflect"
	"testing"
)

func TestNewConfiguration(t *testing.T) {
	defaultConf := NewConfiguration()
	t.Run("Has default values", func(t *testing.T) {
		if defaultConf.dEndpoint == "" {
			t.Error("Expected a default value for dEndpoint, got nil")
		}
		if defaultConf.keyTypes == nil {
			t.Error("Expected a default value for keyTypes, got nil")
		}
		if defaultConf.leServer == "" {
			t.Error("Expected a default value for leServer, got nil")
		}
		if defaultConf.path == "" {
			t.Error("Expected a default value for path, got nil")
		}
	})

	t.Run("Uses env values if available", func(t *testing.T) {
		var c *Configuration

		if os.Setenv("DOCKER_ENDPOINT", "TestEndpointValue") != nil {
			t.Fatal("Unable to set environment variables")
		}
		c = NewConfiguration()
		if c.dEndpoint != "TestEndpointValue" {
			t.Error("Expected TestEndpointValue value for dEndpoint, got ", c.dEndpoint)
		}

		if os.Setenv("LETSENCRYPT_KEYTYPES", "EC256+RSA8192") != nil {
			t.Fatal("Unable to set environment variables")
		}
		c = NewConfiguration()
		if !reflect.DeepEqual(c.keyTypes, []string{"EC256", "RSA8192"}) {
			t.Error("Expected [EC256 RSA8192] value for keyTypes, got ", c.keyTypes)
		}

		if os.Setenv("LETSENCRYPT_SERVER", "TestServerValue") != nil {
			t.Fatal("Unable to set environment variables")
		}
		c = NewConfiguration()
		if c.leServer != "TestServerValue" {
			t.Error("Expected TestServerValue value for leServer, got ", c.leServer)
		}

		if os.Setenv("LETSENCRYPT_PATH", "TestPathValue") != nil {
			t.Fatal("Unable to set environment variables")
		}
		c = NewConfiguration()
		if c.path != "TestPathValue" {
			t.Error("Expected TestPathValue value for dEndpoint, got ", c.path)
		}
	})
}
