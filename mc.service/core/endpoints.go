package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	ex "mc.data/extensions"
)

const (
	DefaultAddr = ":8080"
)

func getHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "43200") // 12 hours in seconds

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// jsonResponse writes a JSON response with the given status code and data
func jsonResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// jsonError writes a JSON error response
func jsonError(w http.ResponseWriter, statusCode int, message string) {
	jsonResponse(w, statusCode, map[string]string{"error": message})
}

func GetHttpServer(sc ServiceContext) *http.Server {
	mux := http.NewServeMux()

	// heartbeat route
	mux.HandleFunc("/api/ping", ping)

	// core functionality routes
	mux.HandleFunc("/api/syncStockData", func(w http.ResponseWriter, r *http.Request) {
		syncStockData(w, r, sc)
	})

	// basic testing routes, will remove eventually
	mux.HandleFunc("/api/test/addByGet", addByGet)
	mux.HandleFunc("/api/test/addByPost", addByPost)

	handler := getHandler(mux)

	return &http.Server{
		Addr:           DefaultAddr,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

// heartbeat routes
func ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "pong"})
}

// core functionalty routes
func GetConfigurationResources(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}

}

type SyncStockDataRequest struct {
	Symbol string `json:"symbol"`
}

func syncStockData(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SyncStockDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Symbol == "" {
		jsonError(w, http.StatusBadRequest, "symbol is required")
		return
	}

	lut, err := sc.SyncSymbolTimeSeriesData(req.Symbol)
	if err != nil {
		if lut.IsZero() {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		jsonResponse(w, http.StatusBadRequest, map[string]any{
			"date":    ex.FmtShort(lut),
			"message": err.Error(),
		})
		return
	}

	// Get the updated metadata to return the last refreshed date
	md, err := sc.PostgresConnection.GetMetaDataBySymbol(sc.Context, req.Symbol)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error getting metadata: %v", err))
		return
	}

	if md == nil {
		jsonError(w, http.StatusInternalServerError, "metadata not found after sync")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"date": ex.FmtShort(md.LastRefreshed)})
}

// Testing endpoints below to ensure functionality
type NumbersToSum struct {
	Number1 int `json:"number1"`
	Number2 int `json:"number2"`
}

// AddByGet adds two numbers via a GET request
func addByGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	number1Str := r.URL.Query().Get("number1")
	number2Str := r.URL.Query().Get("number2")

	number1, err1 := strconv.Atoi(number1Str)
	number2, err2 := strconv.Atoi(number2Str)

	if err1 != nil || err2 != nil {
		jsonError(w, http.StatusBadRequest, "Invalid numbers")
		return
	}

	result := number1 + number2
	jsonResponse(w, http.StatusOK, map[string]int{"result": result})
}

// AddByPost adds two numbers via a POST request
func addByPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var nums NumbersToSum
	if err := json.NewDecoder(r.Body).Decode(&nums); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	result := nums.Number1 + nums.Number2
	jsonResponse(w, http.StatusOK, map[string]int{"result": result})
}
