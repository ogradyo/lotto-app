package tickets

import "time"

type Game string

const (
	GamePowerball   Game = "POWERBALL"
	GameMegaMillions     = "MEGAMILLIONS"
)

type Ticket struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Game       Game      `json:"game"`
	DrawDate   time.Time `json:"draw_date"` // use date-only in DB
	White      [5]int    `json:"white"`
	Special    int       `json:"special"`
	Multiplier *int      `json:"multiplier,omitempty"`
	ImageURL   *string   `json:"image_url,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
