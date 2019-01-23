package go_tool

import (
	"context"
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

var connStrFormat = "amqp://%s:%s@%s:%d/%s"

const(
	ExchangeDirect = amqp.ExchangeDirect
	ExchangeFanout = amqp.ExchangeFanout
	ExchangeHeaders = amqp.ExchangeHeaders
	ExchangeTopic = amqp.ExchangeTopic
)

type MyRabbitMQ struct {
	user        string
	pwd         string
	server      string
	port        int
	virtualHost string

	reConnectDelay time.Duration
	pReSendDelay time.Duration
	pReSendTimes uint32
	pWaitConfirmTimeout time.Duration

	pList map[string]*producer
	cList map[string]*consumer
}

type producer struct {
	Tag string
	ConnStr string
	connection *amqp.Connection
	channel *amqp.Channel
	notifyClose chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	isConnected bool
	reConnectDelay time.Duration
	reSendDelay time.Duration
	reSendTimes uint32
	waitConfirmTimeout time.Duration
	lastConnErr error
	ctx context.Context
	cancelFunc context.CancelFunc
}

type consumer struct {
	Tag string
	QueueName string
	f func(string)
	ConnStr string
	connection *amqp.Connection
	channel *amqp.Channel
	notifyClose chan *amqp.Error
	isConnected bool
	isHandled bool
	reConnectDelay time.Duration
	lastConnErr error
	ctx context.Context
	cancelFunc context.CancelFunc

}

func NewRabbitMQ(user string,pwd string,server string,port int,virtualHost string,
	reConnectDelay time.Duration,pReSendDelay time.Duration,pReSendTimes uint32,pWaitConfirmTimeout time.Duration) *MyRabbitMQ{
	r := &MyRabbitMQ{
		user:user,
		pwd:pwd,
		server:server,
		port:port,
		virtualHost:virtualHost,
		reConnectDelay:reConnectDelay,
		pReSendDelay:pReSendDelay,
		pReSendTimes:pReSendTimes,
		pWaitConfirmTimeout:pWaitConfirmTimeout,
		pList:make(map[string]*producer),
		cList:make(map[string]*consumer),
	}
	return r
}

func (r *MyRabbitMQ) GetConn() (*amqp.Connection,error) {
	conn, err := amqp.Dial(r.getConnStr())
	if err != nil {
		return nil,err
	}
	return conn,nil
}

func (r *MyRabbitMQ)getChannel(connection *amqp.Connection) (*amqp.Channel,error) {
	if connection == nil {
		return nil,errors.New("connection is nil")
	}
	ch,err := connection.Channel()
	if err != nil {
		return nil,err
	}
	return ch,nil
}

func (r *MyRabbitMQ) ExchangeDeclareSimple(connection *amqp.Connection,exchange, exchangeType string) error {
	return r.ExchangeDeclare(connection,exchange,exchangeType,true,false,false,false)
}

func (r *MyRabbitMQ) ExchangeDeclare(connection *amqp.Connection,exchange, exchangeType string, durable, autoDelete, internal, noWait bool) error {
	if connection == nil {
		return errors.New("connection is nil")
	}
	ch,err := r.getChannel(connection)
	if err != nil {
		return err
	}
	defer func(){
		_ = ch.Close()
	}()
	err = ch.ExchangeDeclare(exchange,exchangeType,durable, autoDelete, internal, noWait, nil)
	return err
}

func (r *MyRabbitMQ) ExchangeDelete(connection *amqp.Connection,exchange string,ifUnused, noWait bool) error {
	if connection == nil {
		return errors.New("connection is nil")
	}
	ch,err := r.getChannel(connection)
	if err != nil {
		return err
	}
	defer func(){
		_ = ch.Close()
	}()
	return ch.ExchangeDelete(exchange,false,true)
}

func (r *MyRabbitMQ)QueueDeclareSimple(connection *amqp.Connection,queue string) error{
	return r.QueueDeclare(connection,queue,true,false,false,false)
}

func (r *MyRabbitMQ)QueueDeclare(connection *amqp.Connection,queue string, durable, autoDelete, exclusive, noWait bool) error{
	if connection == nil {
		return errors.New("connection is nil")
	}
	ch,err := r.getChannel(connection)
	if err != nil {
		return err
	}
	defer func(){
		_ = ch.Close()
	}()
	_,err = ch.QueueDeclare(queue,durable, autoDelete, exclusive, noWait,nil)
	return err
}

func (r *MyRabbitMQ)QueueDelete(connection *amqp.Connection,queue string, ifUnused, ifEmpty, noWait bool) error{
	if connection == nil {
		return errors.New("connection is nil")
	}
	ch,err := r.getChannel(connection)
	if err != nil {
		return err
	}
	defer func(){
		_ = ch.Close()
	}()
	_,err = ch.QueueDelete(queue,ifUnused,ifEmpty,noWait)
	return err
}

func (r *MyRabbitMQ)QueueBind(connection *amqp.Connection,queueName, key, exchange string, noWait bool) error {
	if connection == nil {
		return errors.New("connection is nil")
	}
	ch,err := r.getChannel(connection)
	if err != nil {
		return err
	}
	defer func(){
		_ = ch.Close()
	}()
	err = ch.QueueBind(queueName,key,exchange,noWait,nil)
	return err
}

func (r *MyRabbitMQ)QueueUnbind(connection *amqp.Connection,queueName, key, exchange string) error {
	if connection == nil {
		return errors.New("connection is nil")
	}
	ch,err := r.getChannel(connection)
	if err != nil {
		return err
	}
	defer func(){
		_ = ch.Close()
	}()
	err = ch.QueueUnbind(queueName,key,exchange,nil)
	return err
}


func(r *MyRabbitMQ)getConnStr()string{
	return fmt.Sprintf(connStrFormat, r.user, r.pwd, r.server, r.port,r.virtualHost)
}

func(r *MyRabbitMQ)connectTest()(bool,error) {
	conn,err := amqp.Dial(r.getConnStr())
	if err != nil {
		return false,err
	}
	defer func(){
		_ = conn.Close()
	}()
	ch,err := conn.Channel()
	if err != nil {
		return false,err
	}
	defer func(){
		_ = ch.Close()
	}()
	return true,nil
}

func (r *MyRabbitMQ)AddProducer(tag string)error{
	if _,ok := r.pList[tag];ok{
		return errors.New("tag is already exists[" + tag + "]")
	}

	b,err := r.connectTest()
	if !b{
		errMsg := "test connect err:"
		if err != nil {
			return errors.New(errMsg + err.Error())
		} else {
			return errors.New(errMsg + "unknown error")
		}
	}

	p := &producer{
		Tag:tag,
		ConnStr:r.getConnStr(),
		reConnectDelay:r.reConnectDelay,
		reSendDelay:r.pReSendDelay,
		waitConfirmTimeout:r.pWaitConfirmTimeout,
	}
	p.ctx,p.cancelFunc = context.WithCancel(context.Background())
	go p.reConnection()
	r.pList[tag] = p
	return nil
}

func (r *MyRabbitMQ)DelProducer(tag string) {
	if _, ok := r.pList[tag]; ok {
		p := r.pList[tag]
		if p == nil {
			return
		}
		_ = p.Close()
		delete(r.pList, tag)
	}
}

func (r *MyRabbitMQ)Publish(tag string,exchange string, key string, body string) error{
	if _,ok := r.pList[tag];!ok{
		return errors.New("p is not exists[" + tag + "]")
	}
	p := r.pList[tag]
	if p == nil {
		return errors.New("p is nil[" + tag + "]")
	}
	return p.Publish(exchange,key,body)
}

func (r *MyRabbitMQ) Close(){
	for key := range r.pList{
		r.DelProducer(key)
	}
	for key := range r.cList{
		r.DelConsumer(key)
	}
}

func(p *producer)reConnection(){
	for {
		p.isConnected = false
		var err error
		for{
			err = p.connect()
			if err != nil {
				p.lastConnErr = err
				time.Sleep(p.reConnectDelay)
			} else {
				p.lastConnErr = nil
				break
			}
		}
		select {
		case <-p.ctx.Done():
			return
		case <-p.notifyClose:
		}
	}
}

func(p *producer)connect()error{
	conn,err := amqp.Dial(p.ConnStr)
	if err != nil {
		return err
	}
	ch,err := conn.Channel()
	if err != nil {
		return err
	}
	_= ch.Confirm(false)
	p.changeConnection(conn,ch)
	p.isConnected = true
	return nil
}

func(p *producer)changeConnection(connection *amqp.Connection, channel *amqp.Channel){
	p.connection = connection
	p.channel = channel
	p.notifyClose = make(chan *amqp.Error)
	p.notifyConfirm = make(chan amqp.Confirmation)
	p.channel.NotifyClose(p.notifyClose)
	p.channel.NotifyPublish(p.notifyConfirm)
}

func (p *producer) Publish(exchange string, key string, body string) error{
	//if !p.isConnected {
	//	return errors.New("connection is not already")
	//}
	MaxTime := p.reSendTimes
	currTime := uint32(0)
	for {
		err := p.UPublish(exchange,key,body)
		if err != nil {
			currTime++
			if currTime <= MaxTime {
				time.Sleep(p.reSendDelay)
				continue
			} else {
				return err
			}
		}
		break
	}
	ticker := time.NewTicker(p.waitConfirmTimeout)
	select {
		case confirm := <- p.notifyConfirm:
			if confirm.Ack {
				return nil
			}
		case <- ticker.C:
	}
	if p.lastConnErr != nil {
		return p.lastConnErr
	} else {
		return errors.New("wait confirm timeout")
	}

}

func (p *producer) UPublish(exchange string, key string, body string) error{
	//if !p.isConnected {
	//	return errors.New("connection is not already")
	//}
	return p.channel.Publish(
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			Headers:amqp.Table{},
			ContentType:"text/plain",
			ContentEncoding:"",
			Body:[]byte(body),
			DeliveryMode:amqp.Transient,
			Priority:0,
			Timestamp:time.Now(),
		})
}

func (p *producer)Close() error {
	p.cancelFunc()
	if !p.isConnected {
		return nil
	}
	err := p.channel.Close()
	if err != nil {
		return err
	}
	err = p.connection.Close()
	if err != nil {
		return err
	}
	p.isConnected = false
	return nil
}

func (r *MyRabbitMQ)AddConsumer(tag string,queueName string,handler func(string))error{
	if handler == nil {
		return errors.New("handle func can not be nil")
	}
	if _,ok := r.cList[tag];ok{
		return errors.New("tag is already exists[" + tag + "]")
	}
	b,err := r.connectTest()
	if !b{
		errMsg := "test connect err:"
		if err != nil {
			return errors.New(errMsg + err.Error())
		} else {
			return errors.New(errMsg + "unknown error")
		}
	}
	c := &consumer{
		Tag:tag,
		QueueName:queueName,
		f: handler,
		ConnStr:r.getConnStr(),
		reConnectDelay:r.reConnectDelay,
	}
	c.ctx,c.cancelFunc = context.WithCancel(context.Background())
	go c.reConnection()
	go c.reStartHandler()
	r.cList[tag] = c
	return nil
}

func (r *MyRabbitMQ)DelConsumer(tag string){
	if _,ok := r.cList[tag];ok{
		c := r.cList[tag]
		if c == nil {
			return
		}
		_ = c.Close()
		delete(r.cList, tag)
	}
}

func(c *consumer)reConnection(){
	for {
		c.isConnected = false
		var err error
		for{
			err = c.connect()
			if err != nil {
				c.lastConnErr = err
				time.Sleep(c.reConnectDelay)
			} else {
				break
			}
		}
		select {
		case <-c.ctx.Done():
			return
		case <-c.notifyClose:
		}
	}
}

func(c *consumer)connect()error{
	conn,err := amqp.Dial(c.ConnStr)
	if err != nil {
		return err
	}
	ch,err := conn.Channel()
	if err != nil {
		return err
	}
	c.changeConnection(conn,ch)
	b,err := c.handleTest()
	if !b {
		return errors.New("handle test err" + err.Error())
	}
	c.isConnected = true
	c.lastConnErr = nil
	return nil
}

func(c *consumer)changeConnection(connection *amqp.Connection, channel *amqp.Channel){
	c.connection = connection
	c.channel = channel
	c.notifyClose = make(chan *amqp.Error)
	c.channel.NotifyClose(c.notifyClose)
}

func(c *consumer)handleTest()(bool,error){
	err := c.startHandler()
	if err != nil {
		return false,err
	} else {
		return true,nil
	}
}

func(c *consumer)reStartHandler(){
	for{
		c.isHandled = false
		for{
			if !c.isConnected {
				time.Sleep(c.reConnectDelay)
			} else {
				break
			}
		}
		select{
		case <-c.ctx.Done():
			return
		case <-c.notifyClose:
		}
	}
}

func(c *consumer)startHandler()error{
	chanD,err := c.channel.Consume(c.QueueName,c.Tag,true,false,false,false,nil)
	if err != nil {
		return err
	}
	go c.handler(chanD,c.f)
	c.isHandled = true
	return nil
}

func (c *consumer)handler(deliveries <-chan amqp.Delivery,f func(string)){
	for d:= range deliveries {
		f(string(d.Body))
	}
}

func (c *consumer)Close() error {
	c.cancelFunc()
	if !c.isConnected {
		return nil
	}
	err := c.channel.Close()
	if err != nil {
		return err
	}
	err = c.connection.Close()
	if err != nil {
		return err
	}
	c.isConnected = false
	return nil
}