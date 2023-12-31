package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// serveDNS has a 1 in 10 chance to redirect the DNS query to a rickroll. It
// either acts like any other DNS and replies with an IP or just returns a CNAME
// to a rickroll.
func serveDNS(u *net.UDPConn, clientAddr *net.Addr, request *layers.DNS) error {
	var answer string
	var err error

	reply := request
	questionRecord := string(request.Questions[0].Name)

	log.Printf("Resolving %s", questionRecord)

	// 1 in 10 chance to resolve a rickroll
	if n := rand.Intn(10); n == 1 && trap { // if the dice rolls 1 and we're in trap mode
		questionRecord = RICKROLL
		printASCII() // dabbing on the haters
	}

	answer, err = resolveHost(questionRecord)
	if err != nil {
		return err
	}

	replyData, err := DNSreply(reply, answer, questionRecord)
	if err != nil {
		return err
	}
	u.WriteTo(replyData, *clientAddr)
	return nil
}

// DNSreply returns the reply to the query with structured byte data.
func DNSreply(reply *layers.DNS, response, question string) ([]byte, error) {
	var dnsAnswer layers.DNSResourceRecord

	dnsAnswer.Type = layers.DNSTypeA
	dnsAnswer.IP = net.ParseIP(response)
	dnsAnswer.Name = []byte(question)
	dnsAnswer.Class = layers.DNSClassIN

	reply.QR = true
	reply.ANCount = 1
	reply.OpCode = layers.DNSOpCodeQuery
	reply.AA = true
	reply.Answers = append(reply.Answers, dnsAnswer)
	reply.ResponseCode = layers.DNSResponseCodeNoErr

	buf := gopacket.NewSerializeBuffer()
	err := reply.SerializeTo(buf, gopacket.SerializeOptions{})
	if err != nil {
		return nil, fmt.Errorf("error serializing reply: %v", err)
	}

	return buf.Bytes(), nil
}

// resolveHost uses the system DNS service to respond with actual IP address of
// the request host.
func resolveHost(host string) (string, error) {
	resolver := net.Resolver{}
	ips, err := resolver.LookupHost(context.Background(), host)
	if err != nil {
		return "", fmt.Errorf("error resolving host: %w", err)
	}
	return ips[0], nil
}

// sick ASCII
func printASCII() {
	log.Println(RED + `

 #####   ####     ####                       ######    #####    #####   ######    #####   
##   ##   ##       ##                        ##   ##  ##   ##  ##   ##   ##  ##  ##   ##  
##   ##   ##       ##                        ##   ##  ##   ##  ##   ##   ##  ##  ##       
#######   ##       ##                        ######   ##   ##  #######   ##  ##   #####   
##   ##   ##       ##                        ## ##    ##   ##  ##   ##   ##  ##       ##  
##   ##   ##  ##   ##  ##                    ##  ##   ##   ##  ##   ##   ##  ##  ##   ##  
##   ##   ######   ######                    ##   ##   #####   ##   ##  ######    #####   
                                                                                          
####     #######   #####   ######             ######   #####   
 ##       ##  ##  ##   ##   ##  ##              ##    ##   ##  
 ##       ##      ##   ##   ##  ##              ##    ##   ##  
 ##       ####    #######   ##  ##              ##    ##   ##  
 ##       ##      ##   ##   ##  ##              ##    ##   ##  
 ##  ##   ##  ##  ##   ##   ##  ##              ##    ##   ##  
 ######  #######  ##   ##  ######               ##     #####   
                                                               
                           ######    ######   #####    ##  ##  
                           ##   ##     ##    ##   ##   ## ##   
                           ##   ##     ##    ##        ####    
                           ######      ##    ##        ###     
                           ## ##       ##    ##        ####    
                           ##  ##      ##    ##   ##   ## ##   
                           ##   ##   ######   #####    ##  ##  
                                                               
                                    
` + RESET)
}
