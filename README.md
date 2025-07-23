
# Subscription Service

The **Subscription Service** is a RESTful API that allows managing user subscriptions to various services. It is built using Go, Gin, GORM, and PostgreSQL.

## Features

- Create a new subscription
- Retrieve a subscription by ID
- Update or delete a subscription
- List all subscriptions with filters
- Aggregate subscription statistics over time

## Technologies Used

- Go (Golang)
- Gin Web Framework
- GORM (ORM for PostgreSQL)
- PostgreSQL
- Swagger for API documentation

## Requirements

- Go 1.24+
- PostgreSQL 17+
- Docker & Docker Compose (for containerized environment)

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/subscription-service.git
cd subscription-service
```

### 2. Set Environment Variables

Create a `.env` file with the following:

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=test
DB_USER=test_exercise
DB_PASSWORD=secret
```

### 3. Run with Docker

```bash
docker-compose up --build
```

### 4. Run Locally Without Docker

Make sure PostgreSQL is running and configured.

```bash
go mod tidy
go run main.go
```

## API Endpoints

Base URL: `http://localhost:8000`

### Create Subscription

`POST /subscriptions`

**Request Body:**

```json
{
  "service_name": "Netflix",
  "price": 1000,
  "user_id": "UUID",
  "start_date": “07-2025”,
  "end_date": “10-2025”
}

OR

{
“service_name”: “Yandex Plus”,
“price”: 400,
“user_id”: “60601fee-2bf1-4721-ae6f-7636e79a0cba”,
“start_date”: “07-2025”
}
```

### Get Subscription by ID

`GET /subscriptions/{id}`

### Update Subscription

`PUT /subscriptions/{id}`

**Request Body:**

```json
{
  "price": 1200,
  "end_date": “07-2025”
}
```

### Delete Subscription

`DELETE /subscriptions/{id}`

### List Subscriptions

`GET /subscriptions?user_id=UUID&service_name=Spotify`

### Aggregate Subscriptions

`POST /subscriptions/aggregate`

**Request Body:**

```json
{
  "start_date": “07-2025”,
  "end_date": “10-2025”,
  "user_id": "UUID"
}
```

Returns total subscriptions, total price, and service grouping if needed.

## Swagger Documentation

open [`swagger.json`](./swagger.json) in Swagger Editor (https://editor.swagger.io/).

<img width="1920" height="868" alt="image" src="https://github.com/user-attachments/assets/633689ff-ee59-492b-842c-31701d9ca5fe" />

## License

MIT License — free to use and modify.

## Author

Nurdaulet Khaimuldin —(https://github.com/NuKAHAHA/test_exercise)
