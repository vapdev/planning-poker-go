package main

import "net/http"

func enableCors(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Permitir origens específicas
		if origin == "https://planningpoker.digital" || 
		   origin == "https://www.planningpoker.digital" || 
		   origin == "http://localhost:3000" || 
		   origin == "http://127.0.0.1:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Fallback para permitir todas as origens temporariamente
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Origin, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")
		
		// Se não é origem específica conhecida, não permitir credenciais
		if origin == "https://planningpoker.digital" || origin == "https://www.planningpoker.digital" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		fn(w, r)
	}
}
