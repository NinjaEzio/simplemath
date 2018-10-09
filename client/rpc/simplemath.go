package rpc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
	"math/rand"
	pb "simplemath/api"
	"strconv"
	"time"
)

const (
	address         = "localhost:50051"
	timestampFormat = time.StampNano
)

// AuthItem holds the username/password
type AuthItem struct {
	Username string
	Password string
}

// GetRequestMetadata gets the current request metadata
func (a *AuthItem) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"username": a.Username,
		"password": a.Password,
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security
func (a *AuthItem) RequireTransportSecurity() bool {
	return true
}

func getGRPCConn() (conn *grpc.ClientConn, err error) {
	// Setup the username/password
	auth := AuthItem{
		Username: "valineliu",
		Password: "root",
	}
	creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
	return grpc.Dial(address, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&auth))
}

func GreatCommonDivisor(first, second string) {
	conn, err := getGRPCConn()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	a, _ := strconv.ParseInt(first, 10, 32)
	b, _ := strconv.ParseInt(second, 10, 32)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// sending metadata to server: create a new context with some metadata
	ctx = metadata.AppendToOutgoingContext(ctx, "timestamp", time.Now().Format(timestampFormat))
	defer cancel()
	rsp := pb.GCDResponse{}

	var header, trailer metadata.MD
	err = conn.Invoke(ctx, "/api.SimpleMath/GreatCommonDivisor", &pb.GCDRequest{First: int32(a), Second: int32(b)}, &rsp, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		log.Fatalf("could not compute: %v", err)
	}

	// receiving metadata from server: get the Header and Trailer metadata
	if t, ok := header["timestamp"]; ok {
		log.Printf("timestamp from header: ")
		for i, e := range t {
			log.Printf(" %d. %s", i, e)
		}
	}
	if t, ok := trailer["timestamp"]; ok {
		log.Printf("timestamp from trailer: ")
		for i, e := range t {
			log.Printf(" %d. %s", i, e)
		}
	}

	log.Printf("The Greatest Common Divisor of %d and %d is %d", a, b, rsp.Result)
}

func GetFibonacci(count string) {
	conn, err := getGRPCConn()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	num, _ := strconv.ParseInt(count, 10, 32)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client := pb.NewSimpleMathClient(conn)
	stream, err := client.GetFibonacci(ctx, &pb.FibonacciRequest{Count: int32(num)})
	if err != nil {
		log.Fatalf("could not compute: %v", err)
	}

	i := 0
	// receive the results
	for {
		result, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("failed to recv: %v", err)
		}
		log.Printf("#%d: %d\n", i+1, result.Result)
		i++
	}
}

func Statistics(count string) {
	conn, err := getGRPCConn()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewSimpleMathClient(conn)
	stream, err := client.Statistics(context.Background())
	if err != nil {
		log.Fatalf("failed to compute: %v", err)
	}
	num, _ := strconv.ParseInt(count, 10, 32)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var nums []int
	for i := 0; i < int(num); i++ {
		nums = append(nums, r.Intn(100))
	}
	s := ""
	str := ""
	for i := 0; i < int(num); i++ {
		str += s + strconv.Itoa(nums[i])
		s = " "
	}
	log.Printf("Generate numbers: " + str)
	for _, n := range nums {
		if err := stream.Send(&pb.StatisticsRequest{Number: int32(n)}); err != nil {
			log.Fatalf("failed to send: %v", err)
		}
	}
	result, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to recv: %v", err)
	}
	log.Printf("Count: %d\n", result.Count)
	log.Printf("Max: %d\n", result.Maximum)
	log.Printf("Min: %d\n", result.Minimum)
	log.Printf("Avg: %f\n", result.Average)
}

func PrimeFactorization(count string) {
	conn, err := getGRPCConn()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewSimpleMathClient(conn)
	// sending metadata to server: create a new context with some metadata
	ctx := metadata.AppendToOutgoingContext(context.Background(), "timestamp", time.Now().Format(timestampFormat))
	stream, err := client.PrimeFactorization(ctx)
	if err != nil {
		log.Fatalf("failed to compute: %v", err)
	}
	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				break
			}
			if err != nil {
				log.Fatalf("failed to recv: %v", err)
			}
			log.Printf(in.Result)
		}
		// receiving metadata from server: read trailer
		trailer := stream.Trailer()
		if t, ok := trailer["timestamp"]; ok {
			log.Printf("timestamp from trailer: ")
			for i, e := range t {
				log.Printf(" %d. %s", i, e)
			}
		}
	}()

	num, _ := strconv.ParseInt(count, 10, 32)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var nums []int
	for i := 0; i < int(num); i++ {
		nums = append(nums, r.Intn(1000))
	}
	// receiving metadata from server: read header
	header, err := stream.Header()
	if err != nil {
		log.Fatalf("failed to read header from stream: %v", err)
	}
	if t, ok := header["timestamp"]; ok {
		log.Printf("timestamp from header: ")
		for i, e := range t {
			log.Printf(" %d. %s", i, e)
		}
	}
	for _, n := range nums {
		if err := stream.Send(&pb.PrimeFactorizationRequest{Number: int32(n)}); err != nil {
			log.Fatalf("failed to send: %v", err)
		}
		log.Printf("send number: %d", n)
	}
	stream.CloseSend()
	<-waitc
}
