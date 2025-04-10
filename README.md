# geo

Find objects using geocode and addresses

## Try it

1. Clone this repository
2. Check if Docker installed on your system
3. Type `docker compose up`
4. Open your favourite web browser and enter "http://localhost:8080/swagger" in the search box
5. Register on the system using `/register` endpoint
6. Get the authentication token using `/login` endpoint
7. Add the token to `Authorization: Bearer <your_token` HTTP header
8. Enjoy!

## Description

This project serves as a [DaData](https://dadata.ru/api/) API adapter.

API documentation is available at /swagger endpoint.

## Features

- Authentication
- Password encryption with [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- Authentication token blacklist
