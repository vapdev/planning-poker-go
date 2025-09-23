package main

import (
	"net/http"
	"os"
	"strings"
)

func enableCors(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Definir origens permitidas
		allowedOrigins := []string{
			"https://planningpoker.digital",
			"https://www.planningpoker.digital",
			"http://localhost:3000", // Para desenvolvimento local
		}
		
		// Verificar se está em modo de desenvolvimento
		if os.Getenv("NODE_ENV") != "production" {
			allowedOrigins = append(allowedOrigins, 
				"http://localhost:8080",
				"http://127.0.0.1:3000",
				"http://127.0.0.1:8080",
			)
		}
		
		origin := r.Header.Get("Origin")
		
		// Verificar se a origem está na lista de permitidas
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				originAllowed = true
				break
			}
		}
		
		// Se a origem for permitida, defini-la no header
		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin == "" {
			// Para requisições sem Origin header (como algumas requisições diretas)
			w.Header().Set("Access-Control-Allow-Origin", "https://planningpoker.digital")
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Origin, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight por 24 horas
		
		// Headers específicos para WebSocket
		if strings.Contains(r.URL.Path, "/ws/") {
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Origin, X-Requested-With, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Extensions, Sec-WebSocket-Protocol")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		fn(w, r)
	}
}
