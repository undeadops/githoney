package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	libhoney "github.com/honeycombio/libhoney-go"
)

const version = 0.1

type Config struct {
	Port            string
	HoneycombApiKey string
	GitlabAuthToken string
	HoneyEvent      *libhoney.Builder
}

func loggingMiddleware(next http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, next)
}

func main() {
	var config Config
	config.HoneycombApiKey = os.Getenv("HONEYCOMB_API_KEY")
	config.Port = os.Getenv("PORT")
	config.GitlabAuthToken = os.Getenv("GITLAB_AUTH_TOKEN")

	if config.Port == "" {
		config.Port = "5000"
	}

	// basic initialization for honeycomb
	libhConf := libhoney.Config{
		// TODO change to use APIKey
		WriteKey: config.HoneycombApiKey,
		Dataset:  "gitlab",
		Logger:   &libhoney.DefaultLogger{},
	}
	libhoney.Init(libhConf)
	libhoney.AddField("app_version", version)
	libhoney.AddField("app_name", "githoney")

	//Setup Honey Builder Event
	config.HoneyEvent = libhoney.NewBuilder()

	// Setup HTTP Server
	router := config.runServer()

	log.Fatal(http.ListenAndServe(":"+config.Port, router))
	libhoney.Close()
}

func (c *Config) runServer() http.Handler {
	r := mux.NewRouter().StrictSlash(true)
	r.Use(loggingMiddleware)
	r.Use(c.AuthMiddleware)

	r.HandleFunc("/gitlab", c.gitlabWebhook).Methods("POST")
	return r
}

func (c *Config) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Gitlab-Auth")

		if c.GitlabAuthToken == token {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

func (c *Config) gitlabWebhook(w http.ResponseWriter, r *http.Request) {
	// Take event out of post body
	gitlabEvent, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Failed to read post body")
	}

	// TODO: Add some enrichment of the event before forwarding along
	err = c.forwardHoney(gitlabEvent)

	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Forwarding Failed")
	}

	respondWithJSON(w, http.StatusOK, "OK")
}

func (c *Config) forwardHoney(event []byte) error {
	ev := c.HoneyEvent.NewEvent()
	ev.Metadata = fmt.Sprintf("id %d", rand.Intn(20))
	defer ev.Send()
	defer fmt.Printf("Sending event %s\n", ev.Metadata)

	// unmarshal the JSON blob
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(event), &data)
	if err != nil {
		ev.AddField("error", err.Error())
		ev.Send()
		return err
	}

	ev.Timestamp = time.Now()

	ev.Add(data)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
