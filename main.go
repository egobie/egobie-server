package main

import (
    "github.com/eGobie/egobie-server/routes"
)

func main() {
    routes.Serve(":8000")
}
