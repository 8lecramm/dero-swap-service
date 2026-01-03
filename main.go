package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"dero-swap/cfg"
	"dero-swap/coin"

	"github.com/robfig/cron/v3"
)

func Version() {
	log.Println("Dero Swap Service v0.8.7")
	log.Println("https://github.com/8lecramm/dero-swaps")
}

func Init() {

	// create all swap directories
	os.MkdirAll("swaps/active", 0755)
	os.MkdirAll("swaps/expired", 0755)
	os.MkdirAll("swaps/done", 0755)
}

func main() {

	// This is just a dummy yet
	// TODO: store swaps in RAM and save to file after receiving SIGINT and SIGTERM
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	Version()
	Init()

	cfg.LoadConfig()
	if !cfg.CheckConfig() {
		log.Println("Configuration error. Please check config file!")
		os.Exit(1)
	}

	cfg.LoadWallets()
	cfg.LoadFees()
	coin.LockedBalance.LoadActiveSwaps()

	c := cron.New()
	if cfg.Settings.Mode == cfg.SERVER {
		c.AddFunc("@every 5m", UpdateMarkets)
	}
	c.AddFunc("@every 2m", Delay.CheckBackoff)
	c.Start()

	go Swap_Controller()

	Init_Pricing()
	if cfg.Settings.Mode == cfg.SERVER {
		UpdateMarkets()
		StartServer()
	} else {
		StartClient(cfg.Settings.ServerURL)
	}
}
