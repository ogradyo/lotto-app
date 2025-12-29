package httpapi

import (
	"html/template"
	"log"
	"net/http"
	"time"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ogradyo/lotto-app/internal/tickets"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"github.com/google/uuid"

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
	
	// New: photo upload endpoint
	mux.HandleFunc("POST /api/tickets/photo", s.uploadTicketPhotoHandler)

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
func (s *Server) uploadTicketPhotoHandler(w http.ResponseWriter, r *http.Request) {
    const maxUploadSize = 10 << 20 // 10 MB

    r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

    if err := r.ParseMultipartForm(maxUploadSize); err != nil {
        http.Error(w, "file too large or invalid form", http.StatusBadRequest)
        return
    }

    game := r.FormValue("game")
    if game == "" {
        http.Error(w, "game is required", http.StatusBadRequest)
        return
    }

	// didn't use header so removed it
    // file, header, err := r.FormFile("ticket")
    file, _, err := r.FormFile("ticket")
    if err != nil {
        http.Error(w, "ticket file is required", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Basic content-type check
    buf := make([]byte, 512)
    n, _ := file.Read(buf)
    contentType := http.DetectContentType(buf[:n])
    if contentType != "image/jpeg" && contentType != "image/png" {
        http.Error(w, "only JPEG/PNG allowed", http.StatusBadRequest)
        return
    }
    if _, err := file.Seek(0, io.SeekStart); err != nil {
        http.Error(w, "failed to read file", http.StatusInternalServerError)
        return
    }

    // Generate ID and save to disk, e.g. var/ticket-photos/<id>.jpg
    id := uuid.New().String()
    ext := ".jpg" // or derive from contentType/header.Filename
    dir := "var/ticket-photos"
    if err := os.MkdirAll(dir, 0o755); err != nil {
        http.Error(w, "storage error", http.StatusInternalServerError)
        return
    }

    dstPath := filepath.Join(dir, id+ext)
    dst, err := os.Create(dstPath)
    if err != nil {
        http.Error(w, "storage error", http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    if _, err := io.Copy(dst, file); err != nil {
        http.Error(w, "failed to save file", http.StatusInternalServerError)
        return
    }

    // Insert DB row: ticket_photos(id, game, path, uploaded_at, status)
    // status could start as "pending_ocr"

// TODO: persist ticket photo metadata in DB
//	if err := s.store.InsertTicketPhoto(r.Context(), id, game, dstPath); err != nil {
//        http.Error(w, "db error", http.StatusInternalServerError)
//        return
//    }

    w.WriteHeader(http.StatusCreated)
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, `{"id": %q, "game": %q, "status": "pending_ocr"}`, id, game)
}
