package go_tool

import (
	"fmt"
	"github.com/streadway/amqp"
)

var connStrFormat = "amqp://%s:%s@%s:%d/"
var connection *amqp.Connection = nil

type MyRabbitMQ struct {
	User        string
	Pwd         string
	Server      string
	Port        int
	VirtualHost string
}

func (r *MyRabbitMQ) getConn() error {
	conn, err := amqp.Dial(fmt.Sprintf(connStrFormat, r.User, r.Pwd, r.Server, r.Port))
	if err != nil {
		connection = nil
		return err
	}
	connection = conn
	return nil
}
