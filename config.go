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

func (c *Config) Parse() {
	dbPath := flag.String("dbPath", "", "Path to the database")
	compressDir := flag.String("compressDir", "", "Directory to compress")
	maxConcur := flag.Int("maxConcur", 0, "Maximum concurrency")
	human := flag.Bool("human", false, "Human readable format")

	data, err := os.ReadFile("config.toml")
	checkError(err, "reading config file")

	err = toml.Unmarshal(data, &c)
	checkError(err, "parsing config file")

	flag.Parse()

	if *dbPath != "" {
		c.DbPath = *dbPath
	}
	if *compressDir != "" {
		c.Compressdir = *compressDir
	}
	if *maxConcur != 0 {
		c.Maxconcur = *maxConcur
	}
	if *human {
		c.Human = *human
	}
}
func checkError(err error, msg string) {
	if err != nil {
		log.Fatalf("Error %s: %q", msg, err)
		os.Exit(1)
	}
}
