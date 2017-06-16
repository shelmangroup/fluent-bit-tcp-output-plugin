package main

import (
	"C"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/ugorji/go/codec"
)

type TCPConnection struct {
	conn net.Conn
}

var TCP_OUTPUT_HOST string = os.Getenv("TCP_OUTPUT_HOST")
var c TCPConnection

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "out_tcp", "out_tcp GO!")
}

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	return newTCPConnection()
}

func newTCPConnection() int {
	var err error
	c = TCPConnection{}

	c.conn, err = net.Dial("tcp", TCP_OUTPUT_HOST)
	if err != nil {
		fmt.Printf("Failed to start connection: %v\n", err)
		return output.FLB_ERROR
	}
	fmt.Printf("connected to %v\n", c.conn.RemoteAddr())
	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	var h codec.MsgpackHandle

	var b []byte
	var m interface{}
	var err error
	var enc_data []byte

	b = C.GoBytes(data, length)
	dec := codec.NewDecoderBytes(b, &h)

	// Iterate the original MessagePack array
	for {
		// decode the msgpack data
		err = dec.Decode(&m)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Failed to decode msgpack data: %v\n", err)
			return output.FLB_ERROR
		}

		enc_data, err = encode_as_json(m)
		if err != nil {
			fmt.Printf("Failed to encode json data: %v\n", err)
			return output.FLB_ERROR
		}

		if c.conn == nil {
			fmt.Printf("Connection is closed, reconnecting..\n")
			_ = newTCPConnection()
		}

		_, err = fmt.Fprintf(c.conn, "%s\n", jsonPrettyPrint(enc_data))
		if err != nil {
			fmt.Printf("Failed to send data: %v\n", err)
			_ = c.conn.Close()
			c.conn = nil
			return output.FLB_RETRY
		}
	}
	return output.FLB_OK
}
func prepare_data(record interface{}) interface{} {
	// base case,
	// if val is map, return
	r, ok := record.(map[interface{}]interface{})

	if ok != true {
		return record
	}

	newRecord := make(map[string]interface{})
	for k, v := range r {
		key_string := k.(string)
		// convert C-style string to go string, else the JSON encoder will attempt
		// to base64 encode the array
		if val, ok := v.([]byte); ok {
			// if it IS a byte array, make string
			v2 := string(val)
			// add to new record map
			newRecord[key_string] = v2
		} else {
			// if not, recurse to decode interface &
			// add to new record map
			newRecord[key_string] = prepare_data(v)
		}
	}

	return newRecord
}

func encode_as_json(m interface{}) ([]byte, error) {
	slice := reflect.ValueOf(m)
	timestamp := slice.Index(0).Interface().(uint64)
	record := slice.Index(1).Interface()

	type Log struct {
		Time   uint64
		Record interface{}
	}

	log := Log{
		Time:   timestamp,
		Record: prepare_data(record),
	}

	return json.Marshal(log)
}
func jsonPrettyPrint(in []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, in, "", "  ")
	if err != nil {
		return string(in)
	}
	return out.String()
}

func FLBPluginExit() int {
	if err := c.conn.Close(); err != nil {
		fmt.Printf("Failed to close connection: %v", err)
		return output.FLB_ERROR
	}
	return 0
}

func main() {
}
