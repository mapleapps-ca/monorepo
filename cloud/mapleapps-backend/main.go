// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend
package main

import (
	_ "go.uber.org/automaxprocs"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/cmd"
)

func main() {
	cmd.Execute()
}
