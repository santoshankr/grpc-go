/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func buildConfig() *tls.Config {
	rootPem, err := ioutil.ReadFile("../certs/CAcert.pem")
	if err != nil {
		fmt.Printf("Error reading root CA file: %s\n", err)
		return nil
	}
	rootCerts := x509.NewCertPool()
	ok := rootCerts.AppendCertsFromPEM(rootPem)
	if !ok {
		fmt.Printf("Error append root CA cert to pool: %s\n", err)
		return nil
	}

	cert, err := tls.LoadX509KeyPair("../certs/client.crt", "../certs/client.pem")
	if err != nil {
		fmt.Printf("Error loading client certificates: %s\n", err)
		return nil
	}

	ticketCache := tls.NewLRUClientSessionCache(32)
	c := tls.Config{
		Certificates:       []tls.Certificate{cert},
		ServerName:         "server",
		ClientSessionCache: ticketCache,
		RootCAs:            rootCerts,
	}

	return &c
}

func main() {
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	config := buildConfig()

	// Client 1
	connect(config, name)

	// Client 2
	connect(config, name)
}

func connect(config *tls.Config, name string) {
	// Set up a connection to the server.
	creds := credentials.NewTLS(config)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}
