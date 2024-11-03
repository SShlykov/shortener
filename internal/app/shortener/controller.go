package shortener

import (
	"net/http"

	"github.com/go-playground/validator/v10"
)

type shortenerServer interface {
	GetLink(http.ResponseWriter, *http.Request)
	CreateLink(http.ResponseWriter, *http.Request)
}

func Routes(shortenerServer shortenerServer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /link/{linkID}", shortenerServer.GetLink)
	mux.HandleFunc("POST /link", shortenerServer.CreateLink)

	return mux
}

// Server implements shortenerServer
type Server struct {
	validate *validator.Validate
}

func (s *Server) GetLink(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me
	panic("implement me")
}

type LinkDTO struct {
	Key      string `json:"key" validate:"min=6,max=6,alphanum"`
	Original string `json:"url" validate:"required,http_url"`
}

func (s *Server) CreateLink(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me
	// var link LinkDTO
	// err := validate.Struct(link)
	panic("implement me")
}

func NewServer() *Server {
	return &Server{
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (s *Server) Start() error {
	mux := Routes(s)
	return http.ListenAndServe(":8080", mux)
}
