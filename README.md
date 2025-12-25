# GoMonteCarloGo

The purpose of this project was to learn Golang, relearn one of my favorite academic topics, Monte Carlo Simulations, build a Postgresql database from scratch, and design a front end system. It's probably not great code--It's probably the equivalent of "a face a mother can only love," but that's my cross to bear. The structure is loosely set up how I do manage .NET applications professionally, not necessarily how I've seen other Golang based repos. Keeping that constant helped me think through the architecture, but understand I will eventually need to get over long singular files.

To run service:
    cd mc.service
    # create a .env file based on env.example
    # e.g. copy and edit: cp env.example .env
    # then set THIRD_PARTY_API_KEY in .env
    go run main.go

To run web:
    cd mc.web/frontend
    npm start

Environment variables:
    mc.service expects THIRD_PARTY_API_KEY to be set. For local dev, create mc.service/.env (see mc.service/env.example). In production, inject the variable via your hosting environment or a secrets manager.

To run test(s):
    cd *folder with "_test.go" file*
    to run all tests: go test
    to run verbosely: go test -v
    to run single test: go test -run *test func*

If there are updates in other packages, those can be force updated by running:
    cd mc.service
    go get -u mc.data *to directly update the dependency*
    go mod tidy *or whever the consuming model is*

To start postgresql:
    Install via cmd: brew install postgresql@16
    To run via cmd: brew services start postgresql@16
    To stop via cmd: brew services stop postgresql@16

Useful commands:
    go fmt *will format files to go standards*