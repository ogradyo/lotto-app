package tickets

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

type CreateInput struct {
	UserID     int64
	Game       Game
	DrawDate   time.Time
	White      [5]int
	Special    int
	Multiplier *int
	ImageURL   *string
}

func (s *Store) Create(ctx context.Context, in CreateInput) (*Ticket, error) {
	row := s.db.QueryRow(ctx, `
		INSERT INTO tickets (
			user_id, game, draw_date,
			white1, white2, white3, white4, white5,
			special, multiplier, image_url
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, created_at
	`,
		in.UserID, in.Game, in.DrawDate,
		in.White[0], in.White[1], in.White[2], in.White[3], in.White[4],
		in.Special, in.Multiplier, in.ImageURL,
	)

	var t Ticket
	t.UserID = in.UserID
	t.Game = in.Game
	t.DrawDate = in.DrawDate
	t.White = in.White
	t.Special = in.Special
	t.Multiplier = in.Multiplier
	t.ImageURL = in.ImageURL

	if err := row.Scan(&t.ID, &t.CreatedAt); err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) List(ctx context.Context, userID int64) ([]Ticket, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			id,
			user_id,
			game,
			draw_date,
			white1, white2, white3, white4, white5,
			special,
			multiplier,
			image_url,
			created_at
		FROM tickets
		WHERE user_id = $1
		ORDER BY draw_date DESC, id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ts []Ticket
	for rows.Next() {
		var t Ticket
		var w1, w2, w3, w4, w5 int
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Game,
			&t.DrawDate,
			&w1, &w2, &w3, &w4, &w5,
			&t.Special,
			&t.Multiplier,
			&t.ImageURL,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}
		t.White = [5]int{w1, w2, w3, w4, w5}
		ts = append(ts, t)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return ts, nil
}

