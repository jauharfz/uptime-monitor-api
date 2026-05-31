# Uptime Monitor - System Design

## Database Schema

### users
- id SERIAL PK NOT NULL
- username VARCHAR(255) NOT NULL
- password VARCHAR(60) NOT NULL (bcrypt hash)
- email VARCHAR(255) NOT NULL
- created_at TIMESTAMP DEFAULT NOW()
- updated_at TIMESTAMP DEFAULT NOW()

### monitors
- id SERIAL PK NOT NULL
- user_id INT FK NOT NULL
- url VARCHAR(255) NOT NULL
- check_interval INT NOT NULL (seconds)
- last_checked_at TIMESTAMP NULL (tracks when last pinged)
- created_at TIMESTAMP DEFAULT NOW()
- updated_at TIMESTAMP DEFAULT NOW()

### checks
- id SERIAL PK NOT NULL
- monitor_id INT FK NOT NULL
- status_code INT NOT NULL
- response_time INT NOT NULL (milliseconds)
- created_at TIMESTAMP DEFAULT NOW()
- updated_at TIMESTAMP DEFAULT NOW()

## API Endpoints

### Authentication (Public)
- POST /users/register
- POST /users/login

### Monitor Management (Protected, requires Bearer token)
- POST /monitor - create new monitor
- GET /monitor - list all monitors for user
- GET /monitor/{id} - get monitor details
- PATCH /monitor/{id} - update monitor (url, check_interval)
- DELETE /monitor/{id} - delete monitor
- GET /monitor/{id}/checks - get last 50 checks for monitor

## Architecture

### Background Worker
- Runs every 5 seconds
- Fetches monitors due for checking (where last_checked_at is NULL or expired)
- Makes HTTP GET requests (10s timeout) to each URL
- Records status code and response time
- Updates last_checked_at on success

### Server
- HTTP server with graceful shutdown (5s timeout)
- PostgreSQL database with 3s query timeout
- JWT-based auth via Bearer tokens
- Middleware for protected routes

### Configuration
- Database URL from DATABASE_URL env var
- Server port from PORT env var (default: 8080)
- Loads from .env file on startup
