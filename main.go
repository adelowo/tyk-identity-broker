package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/lonelycode/tyk-auth-proxy/backends"
	"github.com/lonelycode/tyk-auth-proxy/tap"
	"github.com/lonelycode/tyk-auth-proxy/tothic"
	"github.com/lonelycode/tyk-auth-proxy/tyk-api"
	"net/http"
	"path"
	"strconv"
)

// AuthConfigStore Is the back end we are storing our configuration files to
var AuthConfigStore tap.AuthRegisterBackend

// IdentityKeyStore keeps a record of identities tied to tokens (if needed)
var IdentityKeyStore tap.AuthRegisterBackend

//  config is the system-wide configuration
var config Configuration

// TykAPIHandler is a global API handler for Tyk, wraps the tyk APi in Go functions
var TykAPIHandler tyk.TykAPI

var log = logrus.New()

// Get our bak end to use, new beack-ends must be registered here
func initBackend(profileBackendConfiguration interface{}, identityBackendConfiguration interface{}) {

	AuthConfigStore = &backends.InMemoryBackend{}
	IdentityKeyStore = &backends.RedisBackend{KeyPrefix: "identity-cache."}

	log.Info("[MAIN] Initialising Profile Configuration Store")
	AuthConfigStore.Init(profileBackendConfiguration)
	log.Info("[MAIN] Initialising Identity Cache")
	IdentityKeyStore.Init(identityBackendConfiguration)
}

func init() {
	log.Info("Tyk Identity Broker v0.1")
	log.Info("Copyright Martin Buhr 2016\n")

	loadConfig("tib.conf", &config)
	initBackend(config.BackEnd.ProfileBackendSettings, config.BackEnd.IdentityBackendSettings)

	TykAPIHandler = config.TykAPISettings

	pDir := path.Join(config.ProfileDir, "profiles.json")
	loaderConf := FileLoaderConf{
		FileName: pDir,
	}

	loader := FileLoader{}
	loader.Init(loaderConf)
	loader.LoadIntoStore(AuthConfigStore)

	tothic.TothErrorHandler = HandleError
}

func main() {
	p := mux.NewRouter()
	p.Handle("/auth/{id}/{provider}/callback", http.HandlerFunc(HandleAuthCallback))
	p.Handle("/auth/{id}/{provider}", http.HandlerFunc(HandleAuth))

	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleGetProfile))).Methods("GET")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleAddProfile))).Methods("POST")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleUpdateProfile))).Methods("PUT")
	p.Handle("/api/profiles/{id}", IsAuthenticated(http.HandlerFunc(HandleDeleteProfile))).Methods("DELETE")

	p.Handle("/api/profiles", IsAuthenticated(http.HandlerFunc(HandleGetProfileList))).Methods("GET")

	listenPort := "3010"
	if config.Port != 0 {
		listenPort = strconv.Itoa(config.Port)
	}

	log.Info("[MAIN] Broker Listening on :", listenPort)

	http.ListenAndServe(":"+listenPort, p)
}
