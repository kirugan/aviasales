package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kirugan/aviasales/proxy/cache"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var httpClient = http.Client{}
var caching cache.Cache
var defaultTimeout time.Duration

func Start(port string, timeout time.Duration) {
	servers := strings.Split(os.Getenv("cache"), ",")
	caching = cache.New(servers, timeout)
	defaultTimeout = timeout

	http.Handle("/v2/places.json", restrictHttpMethod("GET", placesHandler))

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Println(err)
	}
}

func placesHandler(w http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), defaultTimeout)
	defer cancel()

	// future optimization: sort values in the same group, sort groups
	// no dog pile effect code!
	cacheKey := request.URL.RawQuery
	cacheBuf, err := caching.Get(cacheKey)
	if err != nil {
		// caching failures won't destroy the service
		// todo don't log timeout errors & misses
		log.Println(err)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(cacheBuf)

		return
	}

	clientReq, err := createClientRequest(request.URL.RawQuery)
	if err != nil {
		respondError(w, err)
		return
	}

	resp, err := httpClient.Do(clientReq.WithContext(ctx))
	if err != nil {
		respondError(w, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// suppose there won't be errors during copy
		io.Copy(w, resp.Body)
		return
	}

	items, err := createItemsFromRawData(resp.Body)
	if err != nil {
		respondError(w, err)
		return
	}

	buf, err := json.Marshal(&items)
	if err != nil {
		respondError(w, err)
		return
	}

	// no need for cas
	go caching.Set(cacheKey, buf)

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

// respondError log error and return empty response
func respondError(w http.ResponseWriter, err error) {
	log.Println(err)
	// graceful degradation
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`[]`))
}

func createClientRequest(rawQuery string) (*http.Request, error) {
	url := "https://places.aviasales.ru/v2/places.json?" + rawQuery
	return http.NewRequest("GET", url, nil)
}

func restrictHttpMethod(method string, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != method {
			http.Error(writer, fmt.Sprintf("only %s method allowed", method), http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
