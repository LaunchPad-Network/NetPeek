package serverslist

import "github.com/gocarina/gocsv"

type Server struct {
	Id       string `csv:"id"`
	Location string `csv:"name"`
}

func ParseCSV(data string) ([]*Server, error) {
	var servers []*Server
	return servers, gocsv.UnmarshalBytes([]byte(data), &servers)
}
