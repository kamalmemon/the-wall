package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var (
	//go:embed templates/*
	templatesFS embed.FS

	//go:embed static/*
	staticFS embed.FS

	//go:embed db/migrations/*
	migrationsFS embed.FS
)

type Server struct {
	db        *sql.DB
	templates *template.Template
}

type Visitor struct {
	ID            int64
	IPHash        string
	AssignedColor string
}

type GuestbookEntry struct {
	ID      int64
	Name    string
	Message string
	Color   string
}

type PageData struct {
	VisitorNumber int64
	VisitorColor  string
	TotalVisitors int64
	Entries       []GuestbookEntry
	TileColors    []string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "wall.db"
	}

	server, err := NewServer(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("starting server", "addr", ":"+port)
	log.Fatal(http.ListenAndServe(":"+port, server))
}

func NewServer(dbPath string) (*Server, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	s := &Server{db: db}

	if err := s.runMigrations(); err != nil {
		return nil, err
	}

	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	s.templates = tmpl

	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()

	staticHandler := http.FileServer(http.FS(staticFS))
	mux.Handle("GET /static/", staticHandler)
	mux.HandleFunc("GET /", s.HandleWall)
	mux.HandleFunc("POST /api/entry", s.HandleCreateEntry)

	mux.ServeHTTP(w, r)
}

func (s *Server) runMigrations() error {
	entries, err := migrationsFS.ReadDir("db/migrations")
	if err != nil {
		return err
	}

	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		content, err := migrationsFS.ReadFile("db/migrations/" + f)
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(string(content)); err != nil {
			// Ignore errors (table already exists, etc.)
		}
	}
	return nil
}

func (s *Server) HandleWall(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ipHash := hashIP(r.RemoteAddr)
	visitor := s.getOrCreateVisitor(ipHash)

	var totalVisitors int64
	s.db.QueryRow("SELECT COUNT(*) FROM visitors").Scan(&totalVisitors)

	entries := s.getEntries()

	// Get unique colors for mosaic
	colorSet := make(map[string]bool)
	var tileColors []string
	for _, e := range entries {
		if !colorSet[e.Color] {
			colorSet[e.Color] = true
			tileColors = append(tileColors, e.Color)
		}
		if len(tileColors) >= 8 {
			break
		}
	}

	data := PageData{
		VisitorNumber: visitor.ID,
		VisitorColor:  visitor.AssignedColor,
		TotalVisitors: totalVisitors,
		Entries:       entries,
		TileColors:    tileColors,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		slog.Warn("render", "error", err)
	}
}

func (s *Server) HandleCreateEntry(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Message string `json:"message"`
		Color   string `json:"color"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Message == "" || req.Color == "" {
		http.Error(w, "message and color required", http.StatusBadRequest)
		return
	}

	if len(req.Name) > 15 {
		req.Name = req.Name[:15]
	}
	if len(req.Message) > 40 {
		req.Message = req.Message[:40]
	}

	result, err := s.db.Exec(
		"INSERT INTO guestbook_entries (name, message, color) VALUES (?, ?, ?)",
		req.Name, req.Message, req.Color,
	)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":      id,
		"name":    req.Name,
		"message": req.Message,
		"color":   req.Color,
	})
}

func (s *Server) getOrCreateVisitor(ipHash string) Visitor {
	var v Visitor
	err := s.db.QueryRow(
		"SELECT id, ip_hash, assigned_color FROM visitors WHERE ip_hash = ?",
		ipHash,
	).Scan(&v.ID, &v.IPHash, &v.AssignedColor)

	if err == sql.ErrNoRows {
		color := randomColor()
		result, _ := s.db.Exec(
			"INSERT INTO visitors (ip_hash, assigned_color) VALUES (?, ?)",
			ipHash, color,
		)
		v.ID, _ = result.LastInsertId()
		v.IPHash = ipHash
		v.AssignedColor = color
	}

	return v
}

func (s *Server) getEntries() []GuestbookEntry {
	rows, err := s.db.Query(
		"SELECT id, name, message, color FROM guestbook_entries ORDER BY id DESC",
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var entries []GuestbookEntry
	for rows.Next() {
		var e GuestbookEntry
		rows.Scan(&e.ID, &e.Name, &e.Message, &e.Color)
		entries = append(entries, e)
	}
	return entries
}

func hashIP(ip string) string {
	parts := strings.Split(ip, ":")
	if len(parts) > 0 {
		ip = parts[0]
	}
	return fmt.Sprintf("%x", ip)
}

func randomColor() string {
	colors := []string{
		"#ff6b6b", "#ffa726", "#ffee58", "#66bb6a",
		"#42a5f5", "#7e57c2", "#ec407a", "#26a69a",
		"#ff7043", "#8d6e63", "#78909c", "#5c6bc0",
	}
	return colors[rand.Intn(len(colors))]
}
