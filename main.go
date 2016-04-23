package main

import (
    "github.com/egobie/egobie-server/routes"
)

func main() {
    routes.Serve(":8000")
}
