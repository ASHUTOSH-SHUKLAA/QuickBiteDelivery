# QuickBite Delivery API

A RESTful backend service for a food delivery platform built with Go and PostgreSQL. This project demonstrates core backend development skills including database design, REST API implementation, and routing.

## Features
- **Restaurants & Menus**: View available restaurants and their specific menus.
- **Order Management**: Place new orders and retrieve order details.
- **Status Updates**: Update the status of an order (e.g., Pending -> Preparing -> Out for Delivery).

## Technologies Used
- Go (Golang)
- PostgreSQL
- `gorilla/mux` (Routing)
- `lib/pq` (PostgreSQL Driver)

## Setup Instructions

1.  **Database Setup**:
    Ensure you have PostgreSQL running. Create a database named `quickbite`.
    Update the connection string in `main.go` if your username/password differs from the default (`user=postgres password=postgres`).

2.  **Run the Server**:
    ```bash
    go run main.go
    ```
    The server will start on `http://localhost:8080`. The application will automatically create the necessary tables (`restaurants`, `menu_items`, `orders`) and seed some initial data.

## Testing Endpoints
You can use the provided `test.http` file with the VS Code REST Client extension to easily test all endpoints.
