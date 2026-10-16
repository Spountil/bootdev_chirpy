package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	apiCfg := apiConfig{}
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	handlerHealthz := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}
	mux.HandleFunc("GET /api/healthz", handlerHealthz)

	handlerMetrics := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		hits := apiCfg.fileserverHits.Load()
		w.Write([]byte(
			fmt.Sprintf(`
				<html>
					<body>
						<h1>Welcome, Chirpy Admin</h1>
						<p>Chirpy has been visited %d times!</p>
					</body>
				</html>`, hits)))
	}
	mux.HandleFunc("GET /admin/metrics", handlerMetrics)

	handlerReset := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		apiCfg.fileserverHits.Store(0)
	}
	mux.HandleFunc("POST /admin/reset", handlerReset)

	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	http.ListenAndServe(s.Addr, s.Handler)
}
