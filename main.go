package main

import (
	"bufio"
	"encoding/json"
	"net"
	"strings"
	"time"
)

type ActiveClient struct {
	Hosts   []string
	Clients map[string]*Client
}

func NewActiveClient() *ActiveClient {
	return &ActiveClient{
		Hosts:   []string{},
		Clients: map[string]*Client{},
	}
}

type Client struct {
	Conn         net.Conn
	Host         string
	ActiveClient *ActiveClient
	Reader       bufio.Reader
}

func NewClient(conn net.Conn, activeClient *ActiveClient) *Client {
	return &Client{
		Conn:         conn,
		ActiveClient: activeClient,
	}
}

func (activeClient *ActiveClient) NotifyHosts() {
	for _, cl := range activeClient.Clients {
		// cl.Conn.Write(activeClient.Hosts)
		hosts, er := json.Marshal(cl.ActiveClient.Hosts)
		if er != nil {
			return
		}
		_, ers := cl.Conn.Write(hosts)
		if ers != nil {
			// remove(client.ActiveClient.Clients, client)
			delete(cl.ActiveClient.Clients, cl.Host)
			ind := 0
			for in, host := range cl.ActiveClient.Hosts {
				if host == cl.Host {
					ind = in
				}
			}
			// removing the host from the list
			if ind == 0 {
				activeClient.Hosts = activeClient.Hosts[1:]
			} else {
				activeClient.Hosts = append(activeClient.Hosts[:ind], activeClient.Hosts[ind+1:]...)
			}
			activeClient.Hosts = append(activeClient.Hosts[:ind], activeClient.Hosts[ind+1:]...)
			return
		}
	}
}

func (client *Client) Broadcast() {
	defer client.Conn.Close()
	ticker := time.NewTicker(time.Duration(time.Second * 10))
	for {
		select {
		case <-ticker.C:
			{
				hosts, er := json.Marshal(client.ActiveClient.Hosts)
				if er != nil {
					return
				}
				writer := bufio.NewWriter(client.Conn)
				_, ers := writer.Write(append([]byte(hosts), '\n'))
				writer.Flush()
				// client.Conn.Write(hosts)
				if ers != nil {
					// remove(client.ActiveClient.Clients, client)
					delete(client.ActiveClient.Clients, client.Host)
					ind := 0
					for in, host := range client.ActiveClient.Hosts {
						if host == client.Host {
							ind = in
						}
					}
					// removing the host from the list
					if ind == 0 {
						client.ActiveClient.Hosts = client.ActiveClient.Hosts[1:]
					} else {
						client.ActiveClient.Hosts = append(client.ActiveClient.Hosts[:ind], client.ActiveClient.Hosts[ind+1:]...)
					}
					return
				}
			}
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		println("Error While listening the socket ")
		return
	}
	println("Listening ... ")
	activeClient := NewActiveClient()
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			println(err.Error())
			continue
		}
		reader := bufio.NewReader(conn)
		message, _, erro := reader.ReadLine()
		if erro != nil {
			println(erro.Error())
			continue
		}
		client := NewClient(conn, activeClient)
		client.Host = strings.Split(conn.LocalAddr().String(), ":")[0] + string(message)
		activeClient.Clients[client.Host] = client
		activeClient.Hosts = append(activeClient.Hosts, client.Host)
		println(client.Host)
		activeClient.NotifyHosts()
		go client.Broadcast()
	}
}
