package models

import "time"

// models/avail_seat.go

type AvailSeat struct {
	SeatID     string `json:"seat_id"`      // "A1", "B2", format yang diharapkan frontend
	ShowingId  int    `json:"showing_id"`   // now_showing_id
	IsSold     bool   `json:"is_sold"`      // status apakah kursi sudah terjual
	IsLoveNest bool   `json:"is_love_nest"` // apakah kursi love nest (F7-F10)
}

type CreateOrderRequest struct {
	UsersID      int      `json:"users_id" binding:"-"`
	Price        float64  `json:"price" binding:"required,min=0"`
	PaymentID    int      `json:"payment_id" binding:"required"`
	NowShowingID int      `json:"now_showing_id" binding:"required"`
	CinemaID     int      `json:"cinema_id" binding:"-"`
	SeatsMap     []string `json:"seats_map" binding:"required,min=1"`
}

type CreateOrderResponse struct {
	ID        int       `json:"id"`
	UsersID   int       `json:"users_id"`
	Price     float64   `json:"price"`
	QRCode    string    `json:"qr_code"`
	TicketID  int       `json:"ticket_id"`
	SeatsMap  []string  `json:"seats_map"`
	CreatedAt time.Time `json:"created_at"`
}

type Order struct {
	ID           int       `json:"id"`
	UsersID      int       `json:"users_id"`
	Price        float64   `json:"price"`
	PaymentID    int       `json:"payment_id"`
	NowShowingID int       `json:"now_showing_id"`
	IsPaid       bool      `json:"is_paid"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// order history

type Seat struct {
	Row        string `json:"row"`
	SeatNumber int    `json:"seat_number"`
}

type OrderHistory struct {
	Id           int       `json:"id"`
	UsersId      int       `json:"users_id"`
	Price        int       `json:"price"`
	PaymentId    int       `json:"payment_id"`
	IsPaid       bool      `json:"is_paid"`
	CreatedAt    time.Time `json:"created_at"`
	NowShowingId int       `json:"now_showing_id"`
	MovieTitle   string    `json:"movie_title"`
	ShowDate     time.Time `json:"show_date"`
	ShowTime     string    `json:"show_time"`
	CinemaName   string    `json:"cinema_name"`
	Seats        []Seat    `json:"seats"`
	QrCode       string    `json:"qr_code"`
}
