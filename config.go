package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Dbpath      string
	Compressdir string
	Maxconcur   int
	Human       bool
}

func (c *Config) setDefaults() {
	c.Dbpath = "./oxipng.db"
	c.Compressdir = "."
	c.Maxconcur = 8
	c.Human = false
}

func (c *Config) Parse() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error finding executable path: %q", err)
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.toml")

	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("Error reading config file: %q", err)
		}
		err = toml.Unmarshal(data, &c)
		if err != nil {
			log.Fatalf("Error parsing config file: %q", err)
		}
	} else {
		c.setDefaults()
	}

	// override with CLI flags
	flag.StringVar(&c.Dbpath, "dbpath", c.Dbpath, "Path to the database")
	flag.StringVar(&c.Compressdir, "dir", c.Compressdir, "Directory to compress")
	flag.IntVar(&c.Maxconcur, "maxconcur", c.Maxconcur, "Maximum concurrency")
	flag.BoolVar(&c.Human, "human", c.Human, "Human readable format")

	flag.Parse()
}
