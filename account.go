package main

import (
	"crypto"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/xenolf/lego/acme"
)

// Account represents a users local saved credentials
type Account struct {
	Email        string `json:"email"`
	key          crypto.PrivateKey
	Registration *acme.RegistrationResource `json:"registration"`
	conf         *Configuration
}

var accounts map[string]*Account

// GetAccount gets the account associated to an email address
func GetAccount(email string, conf *Configuration) *Account {
	if _, ok := accounts[email]; ok {
		return accounts[email]
	}

	acc := NewAccount(email, conf)
	accounts[email] = acc
	return acc
}

// NewAccount creates a new account for an email address
func NewAccount(email string, conf *Configuration) *Account {
	accKeysPath := conf.AccountKeysPath(email)
	// TODO: move to function in configuration?
	accKeyPath := accKeysPath + string(os.PathSeparator) + email + ".key"
	if err := checkFolder(accKeysPath); err != nil {
		log.WithFields(log.Fields{"err": err, "account": email}).Fatal("Could not check/create account directory")
	}

	var privKey crypto.PrivateKey
	if _, err := os.Stat(accKeyPath); os.IsNotExist(err) {

		log.WithField("account", email).Info("No account key found.")
		log.WithField("account", email).Info("Generating a P384 EC key.")
		privKey, err = generatePrivateKey(accKeyPath)
		if err != nil {
			log.Fatalf("Could not generate RSA private account key for account %s: %v", email, err)
		}

		log.WithField("path", accKeyPath).Info("Saved key")
	} else {
		privKey, err = loadPrivateKey(accKeyPath)
		if err != nil {
			log.Fatalf("Could not load RSA private key from file %s: %v", accKeyPath, err)
		}
	}

	accountFile := path.Join(conf.AccountPath(email), "account.json")
	if _, err := os.Stat(accountFile); os.IsNotExist(err) {
		return &Account{Email: email, key: privKey, conf: conf}
	}

	fileBytes, err := ioutil.ReadFile(accountFile)
	if err != nil {
		log.Fatalf("Could not load file for account %s -> %v", email, err)
	}

	var acc Account
	err = json.Unmarshal(fileBytes, &acc)
	if err != nil {
		log.Fatalf("Could not parse file for account %s -> %v", email, err)
	}

	acc.key = privKey
	acc.conf = conf

	if acc.Registration == nil {
		log.Fatalf("Could not load account for %s. Registration is nil.", email)
	}

	if acc.conf == nil {
		log.Fatalf("Could not load account for %s. Configuration is nil.", email)
	}

	return &acc
}

/** Implementation of the acme.User interface **/

// GetEmail returns the email address for the account
func (a *Account) GetEmail() string {
	return a.Email
}

// GetPrivateKey returns the private RSA account key.
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	return a.key
}

// GetRegistration returns the server registration
func (a *Account) GetRegistration() *acme.RegistrationResource {
	return a.Registration
}

/** End **/

// Save the account to disk
func (a *Account) Save() error {
	jsonBytes, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(
		path.Join(a.conf.AccountPath(a.Email), "account.json"),
		jsonBytes,
		0600,
	)
}
