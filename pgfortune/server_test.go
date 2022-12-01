package main

import (
	"bytes"
	"net"
	"os/exec"
	"testing"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/assert"
)

/*
测试输入从write传进去,传到pipe的另一端conn，随后从write接收响应
服务端从conn接收数据并从，匹配响应的响应再发到conn。
*/
func TestPgStartupmessage(t *testing.T) {
	assert := assert.New(t)
	conn, write := net.Pipe()
	startupmesage := &pgproto3.StartupMessage{
		ProtocolVersion: 196608,
		Parameters: map[string]string{
			"DateStyle":          "ISO",
			"TimeZone":           "Asia/Shanghai",
			"client_encoding":    "UTF8",
			"database":           "postgres",
			"extra_float_digits": "2",
			"user":               "postgres",
		},
	}
	start := startupmesage.Encode(nil)
	backend := &PgFortuneBackend{
		backend: pgproto3.NewBackend(write, write), //io.reader和io.writer
		conn:    write,
		responder: func() ([]byte, error) {
			return exec.Command("sh", "", options.responseCommand).CombinedOutput()
		},
	}
	go func() {
		err := backend.Run()
		if err != nil {
			t.Error("出错了")
		}
	}()
	func() {
		_, err := conn.Write(start)
		if err != nil {
			t.Error("出错了")
		}
	}()
	buff := make([]byte, 16384)  //创建buffer
	buf := bytes.NewBuffer(buff) //初始化buffer
	_, err := conn.Read(buf.Bytes())
	assert.Nil(err)
	front := pgproto3.NewFrontend(buf, nil)
	msg, err := front.Receive()
	assert.Nil(err)
	assert.IsType(&pgproto3.AuthenticationOk{}, msg)
	msg, err = front.Receive()
	assert.Nil(err)
	assert.IsType(&pgproto3.ParameterStatus{}, msg)

}

func TestPgQuery(t *testing.T) {
	assert := assert.New(t)
	conn, write := net.Pipe()
	startupmesage := &pgproto3.StartupMessage{
		ProtocolVersion: 196608,
		Parameters: map[string]string{
			"DateStyle":          "ISO",
			"TimeZone":           "Asia/Shanghai",
			"client_encoding":    "UTF8",
			"database":           "postgres",
			"extra_float_digits": "2",
			"user":               "postgres",
		},
	}
	start := startupmesage.Encode(nil)
	query := &pgproto3.Query{String: "selet * from test"}
	q := query.Encode(nil)
	backend := &PgFortuneBackend{
		backend: pgproto3.NewBackend(write, write), //io.reader和io.writer
		conn:    write,
		responder: func() ([]byte, error) {
			return exec.Command("sh", "", options.responseCommand).CombinedOutput()
		},
	}
	go func() {
		err := backend.Run()
		if err != nil {
			t.Error("出错了")
		}
	}()
	func() {
		_, err := conn.Write(start)
		if err != nil {
			t.Error("出错了")
		}
	}()
	buff := make([]byte, 16384)  //创建buffer
	buf := bytes.NewBuffer(buff) //初始化buffer
	_, err := conn.Read(buf.Bytes())
	assert.Nil(err)
	front := pgproto3.NewFrontend(buf, nil)
	msg, err := front.Receive()
	assert.Nil(err)
	assert.IsType(&pgproto3.AuthenticationOk{}, msg)
	msg, err = front.Receive()
	assert.Nil(err)
	assert.IsType(&pgproto3.ParameterStatus{}, msg)

	func() {
		_, err := conn.Write(q)
		if err != nil {
			t.Error("出错了")
		}
	}()
	_, err = conn.Read(buf.Bytes())
	assert.Nil(err)
	front = pgproto3.NewFrontend(buf, nil)
	msg, err = front.Receive()
	assert.Nil(err)
	assert.IsType(&pgproto3.RowDescription{}, msg)
	msg, err = front.Receive()
	assert.Nil(err)
	assert.IsType(&pgproto3.DataRow{}, msg)
	msg, err = front.Receive()
	assert.Nil(err)
	assert.IsType(&pgproto3.CommandComplete{}, msg)
	msg, err = front.Receive()
	assert.Nil(err)
	assert.IsType(&pgproto3.ReadyForQuery{}, msg)
}
