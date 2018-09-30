package rpc

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"math/rand"
	pb "simplemath/api"
	"strconv"
	"time"
)

const (
	address = "localhost:50051"
)

func getGRPCConn() (conn *grpc.ClientConn, err error) {
	creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
	return grpc.Dial(address, grpc.WithTransportCredentials(creds))
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
	defer cancel()
	rsp := pb.GCDResponse{}
	err = conn.Invoke(ctx, "/api.SimpleMath/GreatCommonDivisor", &pb.GCDRequest{First: int32(a), Second: int32(b)}, &rsp)
	if err != nil {
		log.Fatalf("could not compute: %v", err)
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
	stream, err := client.PrimeFactorization(context.Background())
	if err != nil {
		log.Fatalf("failed to compute: %v", err)
	}
	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("failed to recv: %v", err)
			}
			log.Printf(in.Result)
		}
	}()

	num, _ := strconv.ParseInt(count, 10, 32)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var nums []int
	for i := 0; i < int(num); i++ {
		nums = append(nums, r.Intn(1000))
	}
	for _, n := range nums {
		if err := stream.Send(&pb.PrimeFactorizationRequest{Number: int32(n)}); err != nil {
			log.Fatalf("failed to send: %v", err)
		}
		log.Printf("send number: %d",n)
	}
	stream.CloseSend()
	<-waitc
}
