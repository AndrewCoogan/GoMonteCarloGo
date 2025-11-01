# GoMonteCarloGo

To run service:
    cd mc.service
    # create a .env file based on env.example
    # e.g. copy and edit: cp env.example .env
    # then set THIRD_PARTY_API_KEY in .env
    go run main.go

To run web:
    cd mc.web
    npm start

Environment variables:
    mc.service expects THIRD_PARTY_API_KEY to be set. For local dev, create mc.service/.env (see mc.service/env.example). In production, inject the variable via your hosting environment or a secrets manager.