# 🎬 Tickitz Movie API

Backend project for **Tickitz Movie** built with **Go (Gin Gonic)** as the backend engine.  
This project implements **struct validation**, **JWT authentication**, **argon2 password hashing**, **PostgreSQL** as the main database, **Redis** for caching, and **Swagger** for API documentation.

---

## 📦 Tech Stack

- **Go (Golang)** with Gin
- **PostgreSQL**
- **Redis**
- **JWT**
- **Argon2**
- **Migrate** (Database migration tool)
- **Docker & Docker Compose**
- **Swagger** (via [Swaggo](https://github.com/swaggo/swag))

---

## 🌎 Environment Variables

Copy `.env.example` to `.env` and fill with your configuration:

```env
# Database
DBNAME=<YOUR_DB_NAME>
DBUSER=<YOUR_DB_USER>
DBHOST=<YOUR_DB_HOST>
DBPORT=<YOUR_DB_PORT>
DBPASS=<YOUR_DB_PASS>

# JWT
JWT_SECRET=<YOUR_JWT_SECRET>
JWT_ISSUER=<YOUR_JWT_ISSUER>

# Redis
RDSHOST=<YOUR_REDIS_HOST>
RDSPORT=<YOUR_REDIS_PORT>

🔧 Installation

Clone the project

git clone https://github.com/<your-username>/tickitz-movie.git


Navigate to project directory

cd tickitz-movie


Install dependencies

go mod tidy


Setup your environment

cp .env.example .env


Install migrate for DB migration
Install migrate

Run the DB migration

migrate -database YOUR_DATABASE_URL -path ./db/migrations up


Run the project

go run ./cmd/main.go

🚧 API Documentation
Method	Endpoint	Body	Description
GET	/ping		Connection testing (returns pong)
GET	/movies		Get all movies
GET	/movies/:id		Get movie details by ID
POST	/movies	title, synopsis, etc.	Add new movie (Admin only)
PUT	/movies/:id	title, synopsis, etc.	Update movie (Admin only)
DELETE	/movies/:id		Delete movie (Admin only)
POST	/auth/login	email, password	Login
POST	/auth/register	email, password	Register new user
GET	/users		Get all users (Admin only)
POST	/orders	movie_id, seats, etc.	Create new order
GET	/orders/:id		Get order details

👉 Full API docs available via Swagger at:

http://localhost:8080/swagger/index.html

📄 LICENSE

MIT License

Copyright (c) 2025 Tickitz

📧 Contact Info

Author: Raihan Inkam

Email: your-email@example.com

🎯 Related Project

Tickitz Frontend (React)


Mau saya tambahkan juga **contoh request/response JSON** di bawah tabel API biar gam
```
