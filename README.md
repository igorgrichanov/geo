# geo

Find objects using geocode and addresses

## Try it

1. Clone this repository
2. Check if Docker installed on your system
3. Type `docker compose up`
4. Open your favourite web browser and enter "http://localhost:8080/swagger" in the search box
5. Register on the system using `/register` endpoint
6. Get the authentication token using `/login` endpoint
7. Add the token to `Authorization: Bearer <your_token>` HTTP header
8. Enjoy!

#### /address/geocode
Specify latitude and longitude in the request body
```
{
  "lat": "55.8481373",
  "lng": "37.6414907"
}
```
Get the list of places at that location
```
{
  "addresses": [
    {
      "city": "Москва",
      "street": "Серебрякова",
      "house": "1/2",
      "lat": "55.847447",
      "lon": "37.640803"
    },
    {
      "city": "Москва",
      "street": "Лазоревый",
      "house": "1",
      "lat": "55.848574",
      "lon": "37.640309"
    },
    ...
  ]
}
```

#### /address/search
Specify the location
```
{
  "query": "г Москва, ул Снежная"
}
```

Get the list of addresses located at specified location

```
{
  "addresses": [
    {
      "city": "Москва",
      "street": "Снежная",
      "house": "",
      "lat": "55.852405",
      "lon": "37.646947"
    },
    {
      "city": "Москва",
      "street": "Снежная",
      "house": "1А",
      "lat": "55.846724",
      "lon": "37.639545"
    },
    ...
  ]
}
```

## Description

This project serves as a [DaData](https://dadata.ru/api/) API adapter.

API documentation is available at /swagger endpoint.

## Features

- Authentication
- Passwords encryption with [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- Authentication token blacklist
- Infrastructure layer test coverage 100%
- Query logging




