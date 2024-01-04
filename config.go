package main

import (
	"flag"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	DbPath      string
	Compressdir string
	Maxconcur   int
	Human       bool
}

func (c *Config) setDefaults() {
	c.DbPath = "./oxipng.db"
	c.Compressdir = "."
	c.Maxconcur = 8
	c.Human = false
}

func (c *Config) Parse() {
	dbPath := flag.String("dbPath", "./oxipng.db", "Path to the database")
	compressDir := flag.String("compressDir", ".", "Directory to compress")
	maxConcur := flag.Int("maxConcur", 8, "Maximum concurrency")
	human := flag.Bool("human", false, "Human readable format")

	if _, err := os.Stat("config.toml"); err == nil {
		data, err := os.ReadFile("config.toml")
		if err != nil {
			log.Fatalf("Error reading config file: %q", err)
		}

		err = toml.Unmarshal(data, &c)
		if err != nil {
			log.Fatalf("Error parsing config file: %q", err)
		}
	}
	flag.Parse()

	if *dbPath != "./oxipng.db" {
		c.DbPath = *dbPath
	}
	if *compressDir != "." {
		c.Compressdir = *compressDir
	}
	if *maxConcur != 8 {
		c.Maxconcur = *maxConcur
	}
	if *human {
		c.Human = *human
	}
}
