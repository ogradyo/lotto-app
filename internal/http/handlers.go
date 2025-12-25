package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ogradyo/lotto-app/internal/tickets"
)

type Server struct {
	DB          *pgxpool.Pool
	TicketStore *tickets.Store
}

func NewServer(db *pgxpool.Pool) *Server {
	return &Server{
		DB:          db,
		TicketStore: tickets.NewStore(db),
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", s.handleHealth)

	r.Route("/api", func(r chi.Router) {
		r.Post("/tickets", s.handleCreateTicket)
	})

	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

type createTicketRequest struct {
	Game       string   `json:"game"`
	DrawDate   string   `json:"draw_date"` // YYYY-MM-DD
	White      []int    `json:"white"`
	Special    int      `json:"special"`
	Multiplier *int     `json:"multiplier"`
	ImageURL   *string  `json:"image_url"`
}

func (s *Server) handleCreateTicket(w http.ResponseWriter, r *http.Request) {
	var req createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Game != string(tickets.GamePowerball) && req.Game != string(tickets.GameMegaMillions) {
		http.Error(w, "invalid game", http.StatusBadRequest)
		return
	}
	if len(req.White) != 5 {
		http.Error(w, "white must have 5 numbers", http.StatusBadRequest)
		return
	}

	drawDate, err := time.Parse("2006-01-02", req.DrawDate)
	if err != nil {
		http.Error(w, "invalid draw_date", http.StatusBadRequest)
		return
	}

	var whiteArr [5]int
	copy(whiteArr[:], req.White)

	// TODO: range checking per game (Powerball vs Mega Millions)

	// TEMP: hardcode user_id = 1
	in := tickets.CreateInput{
		UserID:     1,
		Game:       tickets.Game(req.Game),
		DrawDate:   drawDate,
		White:      whiteArr,
		Special:    req.Special,
		Multiplier: req.Multiplier,
		ImageURL:   req.ImageURL,
	}

	t, err := s.TicketStore.Create(context.Background(), in)
	if err != nil {
		http.Error(w, "failed to create ticket", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(t)
}
