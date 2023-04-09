# Flatfeestack API

The Backend 🔙🔚  for the Flatfee(♭💰)stack platform 

## Setup

The setup will soon be dockerized, but for now you can create your own Postgres instance, run the init.sql script and paste the connection string to a `.env` file

```
POSTGRES_URL="postgresql://postgres:password@localhost:5432/flatfeestack?sslmode=disable"
```

## Start

```make && ./api```

## Development

Don't forget to change the openapi schema in `backend.yaml`, if you change the API.


