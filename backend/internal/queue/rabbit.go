package queue

import (
	"encoding/json"
	"github.com/kushagrarai09/flowforge/internal/domain"
	"github.com/rabbitmq/amqp091-go"
)
type Queue struct { Conn *amqp091.Connection; Ch *amqp091.Channel }
const Name="flowforge.execute"
func Connect(url string)(*Queue,error){
	c,e:=amqp091.Dial(url);if e!=nil{return nil,e};ch,e:=c.Channel();if e!=nil{c.Close();return nil,e}
	_,e=ch.QueueDeclare(Name,true,false,false,false,nil);if e!=nil{ch.Close();c.Close();return nil,e};return &Queue{c,ch},nil
}
func(q *Queue)Publish(m domain.ExecuteMessage)error{b,e:=json.Marshal(m);if e!=nil{return e};return q.Ch.Publish("",Name,false,false,amqp091.Publishing{ContentType:"application/json",DeliveryMode:amqp091.Persistent,Body:b})}
func(q *Queue)Consume()(<-chan amqp091.Delivery,error){return q.Ch.Consume(Name,"",false,false,false,false,nil)}
func(q *Queue)Close(){q.Ch.Close();q.Conn.Close()}
