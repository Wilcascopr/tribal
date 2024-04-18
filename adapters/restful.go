package adapters

import (
	"encoding/json"
	"net/http"
	"tribal/internal"
)

func Init() {
	router := newRouter()
	http.ListenAndServe(":8080", router)
}

func newRouter() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/users", handlerGetUsers)
	return router
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handlerGetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	users, err := internal.GenerateUsers(15_000)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: err.Error(),
		})
		return
	}
	println(len(users))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}