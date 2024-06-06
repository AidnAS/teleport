package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/IBM/sarama"
)

func init() {
	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
}

var (
	brokers = flag.String(
		"brokers",
		os.Getenv("KAFKA_PEERS"),
		"The Kafka brokers to connect to, as a comma separated list",
	)
	version  = flag.String("version", sarama.DefaultVersion.String(), "Kafka cluster version")
	certFile = flag.String(
		"certificate",
		"",
		"The optional certificate file for client authentication",
	)
	keyFile = flag.String("key", "", "The optional key file for client authentication")
	caFile  = flag.String(
		"ca",
		"",
		"The optional certificate authority file for TLS client authentication",
	)
	tlsSkipVerify = flag.Bool(
		"tls-skip-verify",
		false,
		"Whether to skip TLS server cert verification",
	)
)

func createTLSConfiguration() (t *tls.Config) {
	t = &tls.Config{
		InsecureSkipVerify: *tlsSkipVerify,
	}

	/*
		encryptedPrivateKey, err := os.ReadFile(*keyFile)
		if err != nil {
			log.Fatal(err)
		}

		privatekey, err := pkcs8.ParsePKCS8PrivateKey(encryptedPrivateKey, []byte("aidn"))
		if err != nil {
			log.Fatal(err)
		}
		encode := base64.RawURLEncoding.EncodeToString

		certBytes, err := os.ReadFile(*certFile)
		if err != nil {
			log.Fatal(err)
		}
	*/

	if *certFile != "" && *keyFile != "" && *caFile != "" {
		cert, err := tls.LoadX509KeyPair(*certFile, *keyFile) // nolint: gosec
		if err != nil {
			log.Fatal(err)
		}

		caCert, err := os.ReadFile(*caFile)
		if err != nil {
			log.Fatal(err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		t = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: *tlsSkipVerify,
		}
	}
	return t
}

func main() {
	flag.Parse()

	if *brokers == "" {
		log.Fatalln("at least one broker is required")
	}
	splitBrokers := strings.Split(*brokers, ",")

	version, err := sarama.ParseKafkaVersion(*version)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	conf := sarama.NewConfig()
	conf.Producer.Retry.Max = 1
	conf.Producer.RequiredAcks = sarama.WaitForAll
	conf.Producer.Return.Successes = true
	conf.Version = version
	conf.ClientID = "ssl_client"
	conf.Metadata.Full = true
	conf.Net.TLS.Enable = true
	conf.Net.TLS.Config = createTLSConfiguration()

	client, err := sarama.NewClusterAdmin(splitBrokers, conf)
	if err != nil {
		log.Fatalf("failed to create cluster admin: %v", err)
	}

	topics, err := client.ListTopics()
	if err != nil {
		log.Fatalf("failed to list topics: %v", err)
	}
	log.Printf("Found %d topics\n", len(topics))

	if len(topics) == 1 {
		if err := client.CreateTopic("test-topic-2", &sarama.TopicDetail{
			NumPartitions: 1, ReplicationFactor: 1}, false); err != nil {
			log.Fatalf("failed to create topic: %v", err)
		}
		log.Printf("created topic")

		topics, err = client.ListTopics()
		if err != nil {
			log.Fatalf("failed to list topics: %v", err)
		}
	}

	log.Printf("Found %d topics\n", len(topics))
}
