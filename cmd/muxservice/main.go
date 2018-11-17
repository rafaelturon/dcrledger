package muxservice

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpcclient"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gorilla/mux"
	"github.com/rafaelturon/decred-pi-wallet/config"
	"github.com/rs/cors"
	"github.com/sec51/twofactor"
	"github.com/urfave/negroni"
)

const (
	otpFileName      = "otp.bin"
	privKeyPath      = "rpc.key"
	pubKeyPath       = "rpc.cert"
	account          = "acccount"
	userName         = "Decred Pi Wallet"
	tokenTimeoutHour = 10
)

var (
	corsArray = []string{"http://localhost"}
	verifyKey *ecdsa.PublicKey
	ecdsaKey  *ecdsa.PrivateKey
	cfg       *config.Config
	client    *rpcclient.Client
	logger    = config.MuxsLog
)

func fatal(err error) {
	if err != nil {
		log.Critical(err)
	}
}

func initKeys() {
	dcrwalletHomeDir := dcrutil.AppDataDir("dcrwallet", false)

	logger.Debugf("Reading private key %s", privKeyPath)
	signBytes, err := ioutil.ReadFile(filepath.Join(dcrwalletHomeDir, privKeyPath))
	fatal(err)

	ecdsaKey, err = jwt.ParseECPrivateKeyFromPEM(signBytes)
	fatal(err)

	logger.Debugf("Reading public key %s", pubKeyPath)
	verifyBytes, err := ioutil.ReadFile(filepath.Join(dcrwalletHomeDir, pubKeyPath))
	fatal(err)

	verifyKey, err = jwt.ParseECPublicKeyFromPEM(verifyBytes)
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

// TicketBuyRequest objct
type TicketBuyRequest struct {
	SpendLimit dcrutil.Amount `json:"spendLimit"`
	NumTickets int            `json:"numTickets"`
	CodeToken  string         `json:"codeToken"`
}

func startServer() {
	router := mux.NewRouter()
	router.HandleFunc("/about", aboutHandler)
	router.HandleFunc("/login", loginHandler)

	// Static route
	sRoutes := mux.NewRouter().PathPrefix("/web").Subrouter().StrictSlash(true)

	// API middleware
	apiRoutes := mux.NewRouter().PathPrefix("/api").Subrouter().StrictSlash(true)
	apiRoutes.HandleFunc("/twofactor", twoFactorHandler).Methods("GET")
	apiRoutes.HandleFunc("/balance", balanceHandler).Methods("GET")
	apiRoutes.HandleFunc("/tickets", ticketsHandler).Methods("GET")
	apiRoutes.HandleFunc("/tickets/buy", ticketsBuyHandler).Methods("POST")
	apiRoutes.HandleFunc("/turnoff/{token_code:[0-9]+}", turnOffDevice).Methods("GET")

	// CORS options
	c := cors.New(cors.Options{
		AllowedOrigins: corsArray,
	})

	// Create static route negroni handler
	router.PathPrefix("/web").Handler(negroni.New(
		negroni.NewStatic(http.Dir(".")),
		negroni.Wrap(sRoutes),
	))

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

func validateToken(codeToken string) error {
	deserializedOTPData, err := ioutil.ReadFile(otpFileName)
	if err != nil {
		fatal(err)
	}
	deserializedOTP, err := twofactor.TOTPFromBytes(deserializedOTPData, userName)
	err = deserializedOTP.Validate(codeToken)

	return err
}

func turnOffDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	codeToken := vars["token_code"]
	err := validateToken(codeToken)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error validating 2FA token")
		logger.Errorf("Error validating 2FA token %v", err)
	} else {
		deviceMessage, err := TurnOffDevice()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error turning off device")
			logger.Errorf("Error turning off device %v", err)
		} else {
			fmt.Fprintln(w, "Device Response: "+deviceMessage)
		}
	}
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Version: " + config.Version()))
}

func twoFactorHandler(w http.ResponseWriter, r *http.Request) {
	qrBytes, err := GetTwoFactorQR(account, userName, otpFileName)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error setting 2FA token - ", err)
		logger.Errorf("Error setting 2FA token %v", err)
	} else {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(qrBytes)))
		if _, err := w.Write(qrBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error getting 2FA")
			logger.Errorf("Error getting 2FA %v", err)
		}
	}
}

func balanceHandler(w http.ResponseWriter, r *http.Request) {
	t, err := GetBalance()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error getting Balance")
		logger.Errorf("Error getting balance %v", err)
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

func ticketsBuyHandler(w http.ResponseWriter, r *http.Request) {
	var ticketBuyRequest TicketBuyRequest
	requestBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(requestBody, &ticketBuyRequest)
	poolFees, err := dcrutil.NewAmount(cfg.PoolFees)
	ticketFee, err := dcrutil.NewAmount(cfg.TicketFee)

	err = validateToken(ticketBuyRequest.CodeToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error validating 2FA token")
		logger.Errorf("Error validating 2FA token %v", err)
	} else {
		hashes, err := BuyTicket(ticketBuyRequest.SpendLimit, cfg.VotingAddress, &ticketBuyRequest.NumTickets, cfg.PoolAddress, &poolFees, &ticketFee, 10)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error buying ticket - ", err)
			logger.Errorf("Error buying ticket %v", err)
		} else {
			fmt.Fprintln(w, "Transaction done with success: "+hashes)
		}
	}

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

	token := jwt.New(jwt.SigningMethodES512)
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

	tokenString, err := token.SignedString(ecdsaKey)

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
func Start(tcfg *config.Config, tclient *rpcclient.Client) {
	cfg = tcfg
	client = tclient
	config.InitLogRotator(cfg.LogFile)
	UseLogger(logger)
	logger.Infof("APIKey %s", cfg.APIKey)
	initKeys()
	startServer()
}
