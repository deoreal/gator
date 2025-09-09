package main

import (
	"fmt"
	"log"
	"os"

	"github.com/deoreal/gator/internal/config"
)

func main() {
	var cfg config.Config
	fmt.Println("This is blogbot")

	cfg.DBURL = "postgres://newDB"
	cfg.CurrentUserName = "mili"

	if err := config.WriteConfig(&cfg); err != nil {
		log.Fatalf("failed to write config: %s", err)
		os.Exit(1)
	}

	cfg, _ = config.ReadConfig()

	fmt.Printf("cfg: %v", cfg)
}
