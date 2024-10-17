package config

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"

	"github.com/caarlos0/env/v6"
)

type FlagRunAddr struct { // host:port for launching the server
	Host string `env:"SERVER_ADDRESS_HOST"`
	Port int    `env:"SERVER_ADDRESS_PORT"`
}

var HostFlags FlagRunAddr
var UrlID string // {id} for shortening url in POST request

func (f FlagRunAddr) String() string {
	return net.JoinHostPort(f.Host, strconv.Itoa(f.Port))
}

func (f *FlagRunAddr) Set(s string) error {
	log.Printf("Setting flag with value: %s", s) // <-- Эта строка
	StrHost, StrPort, err := net.SplitHostPort(s)
	if err != nil {
		return err
	}
	f.Host = StrHost
	f.Port, err = strconv.Atoi(StrPort)
	if err != nil {
		return err
	}
	return nil
}

func ParseFlags() {
	var envErrHostFlags error

	// Parse from the env variables first
	envErrHostFlags = env.Parse(&HostFlags)
	UrlID = os.Getenv("BASE_URL")

	// If no success with env variables then parse from flags
	flag.Var(&HostFlags, "a", "address and port to run server")
	flag.Func("b", "shortened URL path", func(s string) error {
		if !regexp.MustCompile(`[a-zA-Z0-9-]+$`).MatchString(s) {
			return fmt.Errorf("Invalid URL ID: %s", s)
		}
		UrlID = s
		return nil
	})

	if envErrHostFlags != nil || (HostFlags.Host == "" && HostFlags.Port == 0) {
		log.Println("Error parsing host flags: ", envErrHostFlags)
	}
	if UrlID == "" {
		log.Println("Error parsing url ID: ", UrlID)
	}
	if envErrHostFlags != nil || UrlID == "" {
		flag.Parse()
	}
}
