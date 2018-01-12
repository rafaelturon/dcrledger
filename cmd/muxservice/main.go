package muxservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"crypto/rsa"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gorilla/mux"
	"github.com/rafaelturon/decred-pi-wallet/config"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

const (
	privKeyPath      = "app.rsa"
	pubKeyPath       = "app.rsa.pub"
	userName         = "Decred Pi Wallet"
	tokenTimeoutHour = 10
)

var (
	corsArray = []string{"http://localhost"}
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
	cfg       *config.Config
	logger    = config.MuxsLog
)

func fatal(err error) {
	if err != nil {
		log.Critical(err)
	}
}

func initKeys() {
	logger.Debugf("Reading private key %s", privKeyPath)
	signBytes, err := ioutil.ReadFile(privKeyPath)
	fatal(err)

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	fatal(err)

	logger.Debugf("Reading public key %s", pubKeyPath)
	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	fatal(err)

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	fatal(err)
}

// UserCredentials stores data to login
type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// User basic information
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Response API calls
type Response struct {
	Data string `json:"data"`
}

// Token is JWT object string
type Token struct {
	Token string `json:"token"`
}

func startServer() {
	router := mux.NewRouter()
	router.HandleFunc("/about", aboutHandler)
	router.HandleFunc("/login", loginHandler)

	// API middleware
	apiRoutes := mux.NewRouter().PathPrefix("/api").Subrouter().StrictSlash(true)
	apiRoutes.HandleFunc("/balance", balanceHandler)
	apiRoutes.HandleFunc("/tickets", ticketsHandler)

	// CORS options
	c := cors.New(cors.Options{
		AllowedOrigins: corsArray,
	})

	// Create a new negroni for the api middleware
	router.PathPrefix("/api").Handler(negroni.New(
		negroni.HandlerFunc(validateTokenMiddleware),
		negroni.Wrap(apiRoutes),
		c,
	))

	logger.Infof("Listening API at %s", cfg.APIListen)
	// Bind to a port and pass our router in
	logger.Critical(http.ListenAndServe(cfg.APIListen, router))
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Version: " + config.Version()))
}

func balanceHandler(w http.ResponseWriter, r *http.Request) {
	t, err := GetBalance()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error getting Balance")
		logger.Errorf("Error getting balance %v", err)
		fatal(err)
	}
	jsonResponse(t, w)
}

func ticketsHandler(w http.ResponseWriter, r *http.Request) {
	t, err := GetTickets()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error getting Tickets")
		logger.Errorf("Error getting tickets %v", err)
		fatal(err)
	}
	jsonResponse(t, w)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var user UserCredentials

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Error in request")
		logger.Errorf("Error in request %v", err)
		return
	}

	if user.Username != cfg.APIKey || user.Password != cfg.APISecret {
		w.WriteHeader(http.StatusForbidden)
		fmt.Println("Error logging in")
		fmt.Fprint(w, "Invalid credentials")
		logger.Warnf("Invalid credentials %v", err)
		return
	}

	token := jwt.New(jwt.SigningMethodRS256)
	claims := make(jwt.MapClaims)
	claims["admin"] = true
	claims["name"] = userName
	claims["exp"] = time.Now().Add(cfg.APITokenDuration).Unix()
	claims["iat"] = time.Now().Unix()
	token.Claims = claims

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error extracting the key")
		logger.Errorf("Error extracting the key %v", err)
		fatal(err)
	}

	tokenString, err := token.SignedString(signKey)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error while signing the token")
		logger.Errorf("Error while signing the token %v", err)
		fatal(err)
	}

	response := Token{tokenString}
	jsonResponse(response, w)

}

func validateTokenMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			return verifyKey, nil
		})

	if err == nil {
		if token.Valid {
			next(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Token is not valid")
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized access to this resource")
	}

}

func jsonResponse(response interface{}, w http.ResponseWriter) {

	json, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func main() {

}

// Start HTTP request multiplexer service
func Start(tcfg *config.Config) {
	cfg = tcfg
	config.InitLogRotator(cfg.LogFile)
	UseLogger(logger)
	logger.Infof("APIKey %s", cfg.APIKey)
	initKeys()
	startServer()
}
