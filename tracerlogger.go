package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/spagettikod/gotracer"
	"github.com/spagettikod/sdb"
)

const SDBDomainName = "tracerlogger"

var (
	port, accessKey, secretKey string
	db                         sdb.SimpleDB
)

func init() {
	flag.StringVar(&port, "p", "", "COM port where EPsolar Tracer is connected")
	flag.StringVar(&accessKey, "a", "", "AWS Access Key to be able to connect to SimpleDB")
	flag.StringVar(&secretKey, "s", "", "AWS Secret Key, to be able to connect to SimpleDB")
}

func setup() {
	db = sdb.NewSimpleDB(accessKey, secretKey, sdb.SDBRegionEUWest1)
	lsr, err := db.ListDomains()
	if err != nil {
		log.Fatalf("tracelogger.setup - list sdb domains failed: %v", err)
	}
	var exists = false
	for _, d := range lsr.DomainNames {
		exists = d == SDBDomainName
	}
	if !exists {
		_, err = db.CreateDomain(SDBDomainName)
		if err != nil {
			log.Fatalf("tracelogger.setup - create sdb domain failed: %v", err)
		}
	}
}

func doLog() error {
	t, err := gotracer.Status(port)
	if err != nil {
		log.Printf("tracerlogger.doStatus - reading tracer status failed: %v", err)
		return err
	}
	var b []byte
	b, err = json.Marshal(t)
	if err != nil {
		log.Printf("tracerlogger.doStatus - json marshal of read data failed: %v", err)
		return err
	}

	timestamp := time.Now().UTC().Unix()
	item := sdb.NewItem(strconv.FormatInt(timestamp, 10))
	item.AddAttribute("data", string(b))
	_, err = db.PutAttributes(SDBDomainName, item)
	if err != nil {
		log.Printf("tracerlogger.doStatus - putting item into simpledb failed: %v", err)
		return err
	}
	return err
}

func main() {
	flag.Parse()
	if port == "" || accessKey == "" || secretKey == "" {
		flag.PrintDefaults()
		os.Exit(-1)
	}
	setup()
	c := time.Tick(5 * time.Second)
	for _ = range c {
		doLog()
	}
}