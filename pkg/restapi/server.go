package restapi

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/eesrc/geo/pkg/auth"
	"github.com/eesrc/geo/pkg/restapi/validation"
	"github.com/eesrc/geo/pkg/store"
	"github.com/eesrc/geo/pkg/sub/manager"
	"github.com/eesrc/geo/pkg/sub/output"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/acme/autocert"
)

const (
	accessLogFileMode = 0600
)

// Server contains everything needed to run a geoserver
type Server struct {
	params        RestAPIParams
	store         store.Store
	server        *http.Server
	authenticator *auth.Authenticator
	manager       manager.Manager
	done          chan bool
}

// New creates a new HTTP server instance for serving the REST API
func New(params RestAPIParams, manager manager.Manager, store store.Store, authenticator *auth.Authenticator) *Server {
	// TODO(borud) implement access log rotation
	accessLogFile, err := os.OpenFile(params.AccessLog, os.O_CREATE|os.O_APPEND|os.O_WRONLY, accessLogFileMode)
	if err != nil {
		log.Fatalf("Error opening access log '%s': %v", params.AccessLog, err)
	}

	httpServer := Server{
		params:        params,
		store:         store,
		manager:       manager,
		authenticator: authenticator,
	}

	go func() {
		geoSubs, err := getGeoSubscriptionsFromStore(httpServer.store)
		if err != nil {
			// We consider this fatal as the main purpose of the server is to handle subscriptions
			log.Fatalf("Failed to list geo subscriptions: %v", err)
		}
		httpServer.manager.Refresh(geoSubs)
	}()

	// Set up handlers.
	// Order matters here, as some of the handlers either have certain preconditions
	// or directly change the request/response as well as some are reliant on other handlers
	// to be run before others (CORS and proxy headers as ex).

	// Set up handler with logging
	handler := handlers.CombinedLoggingHandler(accessLogFile, httpServer.createRouter())

	// Add recovery
	handler = RecoveryHandler()(handler)

	// Add proxy headers
	handler = handlers.ProxyHeaders(handler)

	// Add CORS headers
	handler = handlers.CORS(
		handlers.AllowedMethods([]string{
			http.MethodGet,
			http.MethodPut,
			http.MethodPost,
			http.MethodDelete,
			http.MethodOptions,
		}),
		handlers.AllowedHeaders([]string{"content-type", "authorization"}),
		handlers.AllowedOrigins([]string{"http://localhost:1234", "https://geo.exploratory.engineering"}),
		handlers.AllowCredentials(),
	)(handler)

	// Compress response based on available compression headers
	handler = handlers.CompressHandler(handler)

	// Populate server
	httpServer.server = &http.Server{
		Handler:      handler,
		Addr:         params.Endpoint,
		WriteTimeout: 180 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return &httpServer
}

// createRouter sets up all the routes for the webserver and restapi.
//
// When maintaining this code, please keep all of the mux
// configuration in one file so that there is one place where
// someone can go to figure out where requests are routed.  Having
// to hunt around the code to figure this out is a pain in the
// ass.
func (s Server) createRouter() *mux.Router {
	r := mux.NewRouter()

	// Set up authentication handlers
	// This will enable authentication endpoints for the different ID providers
	s.addAuthHandlers(r)

	// API paths
	// Generate a subrouter for API which will be used for all subsequent API paths
	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(jsonHeaderMiddleware)
	apiRouter.Use(UUIDWrapper)

	// AuthSessionToUser retrieves the User from the auth session either from an auth provider
	// or token, and makes it available on the request context
	apiRouter.Use(s.authSessionToUserHandlerFunc)

	apiRouter.NotFoundHandler = http.HandlerFunc(notFoundJSON)

	// Profile
	apiRouter.HandleFunc("/profile", s.getUser).Methods("GET")

	// Token management
	apiRouter.HandleFunc("/tokens", s.listTokens).Methods("GET")
	apiRouter.HandleFunc("/tokens", s.createToken).Methods("POST")
	apiRouter.HandleFunc("/tokens/{tokenID}", s.getToken).Methods("GET")
	apiRouter.HandleFunc("/tokens/{tokenID}", s.updateToken).Methods("PUT")
	apiRouter.HandleFunc("/tokens/{tokenID}", s.deleteToken).Methods("DELETE")

	// Collection management
	apiRouter.HandleFunc("/collections", s.listCollections).Methods("GET")
	apiRouter.HandleFunc("/collections", s.createCollection).Methods("POST")
	apiRouter.HandleFunc("/collections/{collectionID}", s.getCollection).Methods("GET")
	apiRouter.HandleFunc("/collections/{collectionID}", s.updateCollection).Methods("PUT")
	apiRouter.HandleFunc("/collections/{collectionID}", s.deleteCollection).Methods("DELETE")
	apiRouter.HandleFunc("/collections/{collectionID}/stream", s.collectionWebsocketData).Methods("GET")

	// Tracker management
	apiRouter.HandleFunc("/collections/{collectionID}/trackers", s.listTrackers).Methods("GET")
	apiRouter.HandleFunc("/collections/{collectionID}/trackers", s.createTracker).Methods("POST")
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}", s.getTracker).Methods("GET")
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}", s.updateTracker).Methods("PUT")
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}", s.deleteTracker).Methods("DELETE")
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}/stream", s.trackerWebsocketData).Methods("GET")

	// Tracker positions
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}/positions", s.listTrackerPositions).Methods("GET")
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}/positions", s.createTrackerPos).Methods("POST")
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}/positions/{positionID}", s.getTrackerPosition).Methods("GET")
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}/positions/{positionID}", s.deleteTrackerPosition).Methods("DELETE")

	// Subscription management
	apiRouter.HandleFunc("/subscriptions", s.listSubscriptions).Methods("GET")
	apiRouter.HandleFunc("/subscriptions", s.createSubscription).Methods("POST")
	apiRouter.HandleFunc("/subscriptions/{subscriptionID}", s.getSubscription).Methods("GET")
	apiRouter.HandleFunc("/subscriptions/{subscriptionID}", s.updateSubscription).Methods("PUT")
	apiRouter.HandleFunc("/subscriptions/{subscriptionID}", s.deleteSubscription).Methods("DELETE")
	apiRouter.HandleFunc("/subscriptions/{subscriptionID}/stream", s.subscriptionWebsocketData).Methods("GET")

	// Subscription management for collection
	apiRouter.HandleFunc("/collections/{collectionID}/subscriptions", s.listCollectionSubscriptions).Methods("GET")
	apiRouter.HandleFunc("/collections/{collectionID}/subscriptions", s.createSubscription).Methods("POST")
	apiRouter.HandleFunc("/collections/{collectionID}/subscriptions/{subscriptionID}", s.getSubscription).Methods("GET")
	apiRouter.HandleFunc("/collections/{collectionID}/subscriptions/{subscriptionID}", s.updateSubscription).Methods("PUT")
	apiRouter.HandleFunc("/collections/{collectionID}/subscriptions/{subscriptionID}", s.deleteSubscription).Methods("DELETE")

	// Subscription management for trackers
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}/subscriptions", s.listTrackerSubscriptions).Methods("GET")
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}/subscriptions", s.createSubscription).Methods("POST")
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}/subscriptions/{subscriptionID}", s.getSubscription).Methods("GET")
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}/subscriptions/{subscriptionID}", s.updateSubscription).Methods("PUT")
	apiRouter.HandleFunc("/collections/{collectionID}/trackers/{trackerID}/subscriptions/{subscriptionID}", s.deleteSubscription).Methods("DELETE")

	// Shapes collection management
	apiRouter.HandleFunc("/shapecollections", s.listShapeCollections).Methods("GET")
	apiRouter.HandleFunc("/shapecollections", s.createShapeCollection).Methods("POST")
	apiRouter.HandleFunc("/shapecollections/{shapeCollectionID}", s.getShapeCollection).Methods("GET")
	apiRouter.HandleFunc("/shapecollections/{shapeCollectionID}", s.updateShapeCollection).Methods("PUT")
	apiRouter.HandleFunc("/shapecollections/{shapeCollectionID}", s.deleteShapeCollection).Methods("DELETE")

	// FeatureCollection operations
	apiRouter.HandleFunc("/shapecollections/{shapeCollectionID}/geojson", s.getFeatureCollection).Methods("GET")
	apiRouter.HandleFunc("/shapecollections/{shapeCollectionID}/geojson", s.updateFeatureCollection).Methods("PUT")
	apiRouter.HandleFunc("/shapecollections/{shapeCollectionID}/geojson", s.createShape).Methods("POST")

	// Single shape operations
	apiRouter.HandleFunc("/shapecollections/{shapeCollectionID}/shapes", s.listShapesByCollection).Methods("GET")
	apiRouter.HandleFunc("/shapecollections/{shapeCollectionID}/shapes/{shapeID}", s.getShape).Methods("GET")
	apiRouter.HandleFunc("/shapecollections/{shapeCollectionID}/shapes/{shapeID}", s.deleteShape).Methods("DELETE")

	apiRouter.HandleFunc("/shapecollections/{shapeCollectionID}/shapes/{shapeID}/geojson", s.getShapeGeoJSON).Methods("GET")
	apiRouter.HandleFunc("/shapecollections/{shapeCollectionID}/shapes/{shapeID}/geojson", s.updateShapeGeoJSON).Methods("PUT")

	// Teams management
	apiRouter.HandleFunc("/teams", s.listTeams).Methods("GET")
	apiRouter.HandleFunc("/teams", s.createTeam).Methods("POST")
	apiRouter.HandleFunc("/teams/accept", notImplemented).Methods("POST")
	apiRouter.HandleFunc("/teams/{teamID}", s.getTeam).Methods("GET")
	apiRouter.HandleFunc("/teams/{teamID}", s.updateTeam).Methods("PUT")
	apiRouter.HandleFunc("/teams/{teamID}", s.deleteTeam).Methods("DELETE")
	apiRouter.HandleFunc("/teams/{teamID}/members", notImplemented).Methods("GET")
	apiRouter.HandleFunc("/teams/{teamID}/members/{userID}", notImplemented).Methods("GET")
	apiRouter.HandleFunc("/teams/{teamID}/members/{userID}", notImplemented).Methods("DELETE")

	// Team invites
	apiRouter.HandleFunc("/teams/{teamID}/invites", notImplemented).Methods("GET")
	apiRouter.HandleFunc("/teams/{teamID}/invites", notImplemented).Methods("POST")
	apiRouter.HandleFunc("/teams/{teamID}/invites/{code}", notImplemented).Methods("GET")
	apiRouter.HandleFunc("/teams/{teamID}/invites/{code}", notImplemented).Methods("DELETE")

	return r
}

func (s *Server) Start() error {
	result := make(chan error)
	go func(result chan error) {
		if s.params.ACME.Enabled {
			log.Info("Using Let's Encrypt for certificates")
			// See https://godoc.org/golang.org/x/crypto/acme/autocert#example-Manager
			m := &autocert.Manager{
				Cache:      autocert.DirCache(s.params.ACME.SecretDir),
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(s.params.ACME.HostList()...),
			}

			// Start a server listening on http for setting up certificates
			go func() {
				if err := http.ListenAndServe(":http", m.HTTPHandler(nil)); err != nil {
					log.Fatal("Failed to start http server for ACME", err)
				}
			}()

			s.server.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}
			s.startServer(result, true, "", "")
			return
		}
		if s.params.TLSKeyFile != "" && s.params.TLSCertFile != "" {
			log.Infof("Using TLS configuration in %s/%s", s.params.TLSCertFile, s.params.TLSKeyFile)
			s.startServer(result, true, s.params.TLSCertFile, s.params.TLSKeyFile)
			return
		}
		s.startServer(result, false, "", "")
	}(result)

	select {
	case err := <-result:
		return err
	case <-time.After(100 * time.Millisecond):
		break
	}
	return nil
}

func (s *Server) Stop() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFunc()

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
	case <-s.done:
	}
	return nil
}

func (s *Server) startServer(result chan error, tls bool, tlscert, tlskey string) {
	defer func() {
		s.done <- true
	}()

	if tls {
		log.Info("REST API runs on port 443")
		s.server.Addr = ":https"
		if err := s.server.ListenAndServeTLS(s.params.TLSCertFile, s.params.TLSKeyFile); err != http.ErrServerClosed {
			result <- err
		}
		return
	}
	log.Infof("REST API runs on %s", s.params.Endpoint)
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		result <- err
	}
}

func getGeoSubscriptionsFromStore(store store.Store) ([]output.GeoSubscription, error) {
	// Populate subscriptions from DB along with shapes.
	geoSubModels, err := store.ListGeoSubscriptions(0, 1000)
	if err != nil {
		return []output.GeoSubscription{}, err
	}

	geoSubs := make([]output.GeoSubscription, len(geoSubModels))

	// Create indexes and initialize new GeoSubscriptions from models
	for i, geoSubModel := range geoSubModels {
		geoSubs[i] = output.NewGeoSubscriptionFromModel(geoSubModel, store)
	}

	return geoSubs, nil
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	validation.NewErrorResponse(http.StatusNotImplemented).WriteHTTPError(w)
}

func notFoundJSON(w http.ResponseWriter, r *http.Request) {
	notFoundResponse := validation.NewErrorResponse(http.StatusNotFound)
	notFoundResponse.Detail = "The path '" + r.URL.Scheme + "://" + r.Host + r.URL.Path + "' does not exist"
	notFoundResponse.WriteHTTPError(w)
}

func jsonHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
