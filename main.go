package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/buaazp/fasthttprouter"
	"github.com/grokify/simplego/encoding/jsonutil"
	"github.com/grokify/simplego/fmt/fmtutil"
	"github.com/grokify/simplego/net/http/httpsimple"
	hum "github.com/grokify/simplego/net/httputilmore"
	"github.com/grokify/simplego/net/urlutil"
	"github.com/grokify/simplego/strconv/strconvutil"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	rc "github.com/grokify/go-ringcentral-client/office/v1/client"
	rcu "github.com/grokify/go-ringcentral-client/office/v1/util"
	rco "github.com/grokify/oauth2more/ringcentral"
)

const (
	DefaultPort              = "8080"
	DefaultPortInt           = 8080
	ValidationTokenHeader    = "Validation-Token"
	MessageStoreEventFilter  = "/restapi/v1.0/account/~/extension/~/message-store"
	SMSEventFilter           = "/restapi/v1.0/account/~/extension/~/message-store/instant?type=SMS"
	RenewalEventFilterFormat = "/restapi/v1.0/subscription/%s?threshold=%d&interval=%d"
)

var (
	OutboundWebhookUrl           = ""                                  // Simple inbound webhook like Zapier or Chathooks
	InboundWebhookUrl            = "https://12345678.ngrok.io/webhook" // Server URL the RingCentral API will send to
	CurrentWebhookSubscriptionId = ""                                  // Current SubscriptionID to renew
	ExpiresIn                    = 60 * 60 * 24 * 7
	RenewalThresholdTime         = 60 * 60 * 24
	RenewalIntervalTime          = 60 * 60
)

// EventFilters determines the RingCentral events this service will subscribe to.
// Threshold is the threshold time (in seconds) remaining before subscription expiration when server should start to send renewal reminder notifications. This time is approximate. It cannot be less than the interval of reminder job execution. It also cannot be greater than a half of this subscription TTL.
// Interval is the interval (in seconds) between reminder notifications. This time is approximate. It cannot be less than the interval of reminder job execution. It also cannot be greater than a half of threshold value.
var RenewalEventFilter = getRenewalEventFilter("~", RenewalThresholdTime, RenewalIntervalTime)
var EventFilters = []string{SMSEventFilter, MessageStoreEventFilter, RenewalEventFilter}

func getRenewalEventFilter(subscriptionID string, threshold, interval int) string {
	return fmt.Sprintf(RenewalEventFilterFormat, subscriptionID, threshold, interval)
}

func setEventFilters() {
	RenewalEventFilter = getRenewalEventFilter("~", RenewalThresholdTime, RenewalIntervalTime)
	EventFilters = []string{SMSEventFilter, MessageStoreEventFilter, RenewalEventFilter}
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Handling webhook...")
	// Check to see if ValidationToken is present. If so, return
	// it immediately.
	validationToken := r.Header.Get(ValidationTokenHeader)
	if len(validationToken) > 0 {
		log.Printf("%s: %s", ValidationTokenHeader, validationToken)
		w.Header().Set(ValidationTokenHeader, validationToken)
		return
	}

	// Read the body to check if this is a renewal event
	httpBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("FAIL - error reading webhook request message body.")
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	log.Debug().
		Str("request_body", string(httpBody)).
		Msg("Receiving Webhook....")

	event := &rcu.Event{}
	err = json.Unmarshal(httpBody, event)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("FAIL - JSON Unmarshal error")
		return
	}

	// If this is renewal event, renew the webhook and return.
	if event.Event == RenewalEventFilter {
		_, err := renewWebhook(event.SubscriptionId)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Error reading body")
			http.Error(w, "can't read body", http.StatusBadRequest)
		}
		return
	}

	if 1 == 1 {
		evt, err := rcu.EventParseBytes(httpBody)
		if err != nil {
			panic(err)
		}
		fmtutil.PrintJSON(evt)
		if evt.IsEventType(rcu.InstantMessageEvent) {
			body, err := evt.GetInstantMessageBody()
			if err != nil {
				panic(err)
			}
			fmtutil.PrintJSON(body)
		}
	}

	// Forward the body to the Webhook URL
	resp, err := http.Post(
		OutboundWebhookUrl,
		hum.ContentTypeAppJsonUtf8,
		bytes.NewBuffer(httpBody))
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Downstream webhook error")
		return
	} else if resp.StatusCode >= 300 {
		log.Warn().
			Int("status_code", resp.StatusCode).
			Msg("Downstream webhook error")
		return
	}
}

func createhookHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := createWebhook()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("FAIL - create webhook")
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("FAIL - parse create webhook response body.")
		return
	}

	w.Header().Set(hum.HeaderContentType, hum.ContentTypeAppJsonUtf8)
	w.Write(body)
}

func renewhookHandler(w http.ResponseWriter, r *http.Request) {
	_, err := renewWebhook()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("FAIL - renew webhook.")
	}
}

func createWebhook() (*http.Response, error) {
	log.Info().Msg("Creating Hook...")
	apiClient, err := newRingCentralClient()
	if err != nil {
		return nil, err
	}

	req := rc.CreateSubscriptionRequest{
		EventFilters: EventFilters,
		DeliveryMode: rc.NotificationDeliveryModeRequest{
			TransportType: "WebHook",
			Address:       InboundWebhookUrl,
		},
		ExpiresIn: int32(ExpiresIn),
	}
	log.Debug().
		Str("body", jsonutil.MustMarshalString(req, true)).
		Msg("create_subscription_request_body")

	return handleWebhookResponse(
		apiClient.PushNotificationsApi.CreateSubscription(
			context.Background(),
			req,
		),
	)
}

func renewWebhook(subscriptionIds ...string) (*http.Response, error) {
	subscriptionId := CurrentWebhookSubscriptionId
	if len(subscriptionIds) > 0 {
		subscriptionId = subscriptionIds[0]
	}
	log.Info().
		Str("hook_subscription_id", subscriptionId).
		Msg("Renewing Webhook")

	apiClient, err := newRingCentralClient()
	if err != nil {
		log.Warn().Err(err).Msg("RENEW NEW RC CLIENT ERROR")
		return nil, err
	}

	return handleWebhookResponse(
		apiClient.PushNotificationsApi.RenewSubscription(
			context.Background(),
			subscriptionId,
		),
	)
}

func handleInternalServerError(w http.ResponseWriter, logmessage string) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	log.Warn().Msg(logmessage)
}

func listhooksHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Getting Subscription List...")
	apiClient, err := newRingCentralClient()
	if err != nil {
		handleInternalServerError(w, fmt.Sprintf("Listhooks: Error getting RC Client: %v", err.Error()))
		return
	}
	info, resp, err := apiClient.PushNotificationsApi.GetSubscriptions(context.Background())
	if err != nil {
		handleInternalServerError(w, fmt.Sprintf("Error calling GetSubscriptions API: %v", err.Error()))
		return
	} else if resp.StatusCode >= 300 {
		handleInternalServerError(w, fmt.Sprintf("Error calling GetSubscriptions API: Status %v", resp.StatusCode))
		return
	}
	bytes, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		handleInternalServerError(w, fmt.Sprintf("Error calling GetSubscriptions API: ReadBody %v", err.Error()))
		return
	}
	w.Header().Set(hum.HeaderContentType, hum.ContentTypeAppJsonUtf8)
	w.Write(bytes)
}

func handleWebhookResponse(info rc.SubscriptionInfo, resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return resp, fmt.Errorf("%v: %v", "API Response Err", err.Error())
	} else if resp.StatusCode > 299 {
		return resp, fmt.Errorf("RingCentral Subscription API request failure status code: %v", resp.StatusCode)
	}

	CurrentWebhookSubscriptionId = info.Id
	log.Info().
		Str("hook_subscription_id", CurrentWebhookSubscriptionId).
		Msg("Created/renewed Webhook")
	return resp, nil
}

func newRingCentralClient() (*rc.APIClient, error) {
	return rcu.NewApiClientPassword(
		rco.ApplicationCredentials{
			ServerURL:    os.Getenv("RINGCENTRAL_SERVER_URL"),
			ClientID:     os.Getenv("RINGCENTRAL_CLIENT_ID"),
			ClientSecret: os.Getenv("RINGCENTRAL_CLIENT_SECRET"),
			AppName:      "github.com/grokify/ringcentral-permahooks",
			AppVersion:   "0.0.1",
		},
		rco.PasswordCredentials{
			Username:  os.Getenv("RINGCENTRAL_USERNAME"),
			Extension: os.Getenv("RINGCENTRAL_EXTENSION"),
			Password:  os.Getenv("RINGCENTRAL_PASSWORD"),
		},
	)
}

func loadEnv() error {
	envPaths := []string{}
	if len(os.Getenv("ENV_PATH")) > 0 {
		envPaths = append(envPaths, os.Getenv("ENV_PATH"))
	}
	return godotenv.Load(envPaths...)
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

type Server struct {
	Port       int
	HTTPEngine string
	Testing    bool
}

func NewServer() Server {
	svr := Server{
		Port:       strconvutil.AtoiOrDefault(os.Getenv("PORT"), DefaultPortInt),
		HTTPEngine: os.Getenv("HTTP_ENGINE"),
	}
	return svr
}

func (svr Server) PortInt() int                       { return svr.Port }
func (svr Server) HttpEngine() string                 { return svr.HTTPEngine }
func (svr Server) RouterFast() *fasthttprouter.Router { return nil }

func (svr Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/webhook", http.HandlerFunc(webhookHandler))
	mux.Handle("/webhook/", http.HandlerFunc(webhookHandler))
	mux.Handle("/createhook", http.HandlerFunc(createhookHandler))
	mux.Handle("/createhook/", http.HandlerFunc(createhookHandler))
	mux.Handle("/renewhook", http.HandlerFunc(renewhookHandler))
	mux.Handle("/renewhook/", http.HandlerFunc(renewhookHandler))
	if svr.Testing {
		mux.Handle("/listhooks", http.HandlerFunc(listhooksHandler))
	}
	return mux
}

func main() {
	err := loadEnv()
	if err != nil {
		log.Fatal().Err(err)
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	InboundWebhookUrl = strings.TrimSpace(os.Getenv("PERMAHOOKS_INBOUND_WEBHOOK_URL"))
	OutboundWebhookUrl = strings.TrimSpace(os.Getenv("PERMAHOOKS_OUTBOUND_WEBHOOK_URL"))

	urlValidator := urlutil.URLValidator{RequiredSchemes: map[string]int{"https": 1}}
	_, err = urlValidator.ValidateURLString(InboundWebhookUrl)
	if err != nil {
		log.Fatal().Err(err).
			Str("environmentVariable", "PERMAHOOKS_INBOUND_WEBHOOK_URL").
			Msg("Environment Variable Error")
	}
	_, err = urlValidator.ValidateURLString(OutboundWebhookUrl)
	if err != nil {
		log.Fatal().Err(err).
			Str("environmentVariable", "PERMAHOOKS_OUTBOUND_WEBHOOK_URL").
			Msg("Environment Variable Error")
	}

	svr := NewServer()

	testing := true // to verify if renewal is working.

	if testing {
		svr.Testing = testing
		ExpiresIn = 180
		RenewalThresholdTime = 80
		RenewalIntervalTime = 30
	}
	setEventFilters()

	if 1 == 0 {
		// Check PORT env. This environment variable name is hard coded to work
		// with Heroku which will auto-assign a port using this name
		port := os.Getenv("PORT")
		if len(strings.TrimSpace(port)) == 0 {
			port = DefaultPort
		}

		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatal().Err(err)
		}

		done := make(chan bool)
		go http.Serve(listener, svr.Router())
		log.Info().Str("port", port).Msg("Server listening")
		<-done
	}

	done := make(chan bool)
	go httpsimple.Serve(svr)
	log.Info().Int("port", svr.Port).Msg("Server listening")
	<-done
}
