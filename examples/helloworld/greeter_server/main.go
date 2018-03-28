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

//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		fmt.Println("Could not get peer from context")
	}
	tlsInfo := (peer.AuthInfo).(credentials.TLSInfo)
	didResume := tlsInfo.State.DidResume
	return &pb.HelloReply{Message: in.Name + " resumed connection? " + strconv.FormatBool(didResume)}, nil
}

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

	cert, err := tls.LoadX509KeyPair("../certs/server.crt", "../certs/server.pem")
	if err != nil {
		fmt.Printf("Error loading server certificates: %s\n", err)
		return nil
	}

	var sessionTicketKey [32]byte
	copy(sessionTicketKey[:], "abcdefghijklmnopqrstuvwxyz012345")
	c := tls.Config{
		Certificates:     []tls.Certificate{cert},
		ClientCAs:        rootCerts,
		ClientAuth:       tls.VerifyClientCertIfGiven,
		SessionTicketKey: sessionTicketKey,
	}

	return &c
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	config := buildConfig()
	creds := credentials.NewTLS(config)

	s := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterGreeterServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
