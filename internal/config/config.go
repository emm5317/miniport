package config

import "os"

type Config struct {
	Host string
	Port string
}

func Load() Config {
	c := Config{Host: "127.0.0.1", Port: "8092"}
	if h := os.Getenv("MINIPORT_HOST"); h != "" {
		c.Host = h
	}
	if p := os.Getenv("MINIPORT_PORT"); p != "" {
		c.Port = p
	}
	return c
}
