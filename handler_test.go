package delta

import (
	"fmt"
	. "github.com/r7kamura/gospel"
	"github.com/r7kamura/router"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupServer() *Server {
	server := NewServer("0.0.0.0", 8484)

	server.AddMasterBackend("production", "0.0.0.0", 8080)
	server.AddBackend("testing", "0.0.0.0", 8081)

	server.OnSelectBackend(func(req *http.Request) []string {
		if req.Method == "GET" {
			return []string{"production", "testing"}
		} else {
			return []string{"production"}
		}
	})

	server.OnMungeHeader(func(backend string, header *http.Header) {
		if backend == "testing" {
			header.Add("X-Delta-Sandbox", "1")
		}
	})

	server.OnBackendFinished(func(responses map[string]*Response) {
	})

	return server
}

func launchBackend(backend string, addr string) {
	router := router.NewRouter()

	router.Get("/", http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(writer, "%s", backend)
	}))

	server := &http.Server{Addr: addr, Handler: router}
	server.ListenAndServe()
}

func get(handler http.Handler, path string) *httptest.ResponseRecorder {
	return request(handler, "GET", path)
}

func request(handler http.Handler, method, path string) *httptest.ResponseRecorder {
	request, _ := http.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestHandler(t *testing.T) {
	go launchBackend("production", ":8080")
	go launchBackend("testing", ":8081")
	handler := NewHandler(setupServer())

	Describe(t, "ServeHTTP", func() {
		Context("request to normal path", func() {
			response := get(handler, "/")

			It("should record only master's response", func() {
				Expect(response.Body.String()).To(Equal, "production")
			})
		})
	})
}
