package main

import (
	"github.com/cristiano-pacheco/pingo/cmd"
)

// @title           Pingo API
// @version         1.0
// @description     Pingo API

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token in the format **Bearer <token>**

// @BasePath  /
func main() {
	cmd.Execute()
}
