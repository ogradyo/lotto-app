package httpapi

import (
	"html/template"
	"log"
	"net/http"
	"time"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ogradyo/lotto-app/internal/tickets"
)

// Server struct
type Server struct {
	db        *pgxpool.Pool
	templates *template.Template
	tickets   *tickets.Store
}

// Constructor
func NewServer(db *pgxpool.Pool, templates *template.Template) *Server {
	return &Server{
		db:        db,
		templates: templates,
		tickets:   tickets.NewStore(db),
	}
}

// Router wiring all handlers
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /api/tickets", s.handleCreateTicketJSON)
	mux.HandleFunc("GET /tickets/new", s.handleShowAddTicketForm)
	mux.HandleFunc("POST /tickets", s.handleCreateTicketForm)

	return mux
}

// Simple health handler
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// Existing JSON ticket handler (your code goes here)
func (s *Server) handleCreateTicketJSON(w http.ResponseWriter, r *http.Request) {
	// move your existing POST /api/tickets JSON logic here
}

// Show the HTML form
func (s *Server) handleShowAddTicketForm(w http.ResponseWriter, r *http.Request) {
	if err := s.templates.ExecuteTemplate(w, "add_ticket.html", nil); err != nil {
		log.Printf("execute template add_ticket.html: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}

// Handle POST /tickets from the form
func (s *Server) handleCreateTicketForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("parse form: %v", err)
		http.Error(w, "bad form data", http.StatusBadRequest)
		return
	}

	gameStr := r.FormValue("game")
	drawDateStr := r.FormValue("draw_date")

	whiteStrs := []string{
		r.FormValue("white1"),
		r.FormValue("white2"),
		r.FormValue("white3"),
		r.FormValue("white4"),
		r.FormValue("white5"),
	}
	specialStr := r.FormValue("special")

	if gameStr == "" || drawDateStr == "" || specialStr == "" {
		s.renderAddTicketForm(w, map[string]any{
			"Error": "All fields are required.",
		})
		return
	}

	drawDate, err := time.Parse("2006-01-02", drawDateStr)
	if err != nil {
		s.renderAddTicketForm(w, map[string]any{
			"Error": "Invalid draw date.",
		})
		return
	}

	var white [5]int
	for i, v := range whiteStrs {
		n, err := strconv.Atoi(v)
		if err != nil {
			s.renderAddTicketForm(w, map[string]any{
				"Error": "White ball numbers must be integers.",
			})
			return
		}
		white[i] = n
	}

	special, err := strconv.Atoi(specialStr)
	if err != nil {
		s.renderAddTicketForm(w, map[string]any{
			"Error": "Special ball must be an integer.",
		})
		return
	}

	// Cast string to tickets.Game
	game := tickets.Game(gameStr)

	input := tickets.CreateInput{
		UserID:   1,           // TODO: real user later. TEMPORARY
		Game:     game,
		DrawDate: drawDate,
		White:    white,
		Special:  special,
		// Multiplier: nil,
		// ImageURL:   nil,
	}

	if _, err := s.tickets.Create(r.Context(), input); err != nil {
		log.Printf("create ticket: %v", err)
		s.renderAddTicketForm(w, map[string]any{
			"Error": "Could not create ticket.",
		})
		return
	}

	s.renderAddTicketForm(w, map[string]any{
		"Success": true,
	})
}

func (s *Server) renderAddTicketForm(w http.ResponseWriter, data any) {
	if err := s.templates.ExecuteTemplate(w, "add_ticket.html", data); err != nil {
		log.Printf("execute template add_ticket.html: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}
