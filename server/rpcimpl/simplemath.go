package rpcimpl

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
	pb "simplemath/api"
	"strconv"
	"time"
)

const (
	timestampFormat = time.StampNano
)

type SimpleMathServer struct{}

func (sms *SimpleMathServer) GreatCommonDivisor(ctx context.Context, in *pb.GCDRequest) (*pb.GCDResponse, error) {
	// sending metadata to client: create trailer, using defer to record timestamp of function return
	defer func() {
		trailer := metadata.Pairs("timestamp", time.Now().Format(timestampFormat))
		grpc.SetTrailer(ctx, trailer)
	}()

	// receiving metadata from client: get metadata from context
	md, _ := metadata.FromIncomingContext(ctx)
	if t, ok := md["timestamp"]; ok {
		log.Printf("timestamp from metadata: ")
		for i, e := range t {
			log.Printf(" %d. %s", i, e)
		}
	}

	first := in.First
	second := in.Second
	for second != 0 {
		first, second = second, first%second
	}
	// sending metadata to client: create and send header
	header := metadata.New(map[string]string{"timestamp": time.Now().Format(timestampFormat)})
	grpc.SendHeader(ctx, header)
	return &pb.GCDResponse{Result: first}, nil
}

func (sms *SimpleMathServer) GetFibonacci(in *pb.FibonacciRequest, stream pb.SimpleMath_GetFibonacciServer) error {
	a, b := 0, 1
	for i := 0; i < int(in.Count); i++ {
		stream.Send(&pb.FibonacciResponse{Result: int32(a)})
		a, b = b, a+b
	}
	return nil
}

func (sms *SimpleMathServer) Statistics(stream pb.SimpleMath_StatisticsServer) error {
	var count, maximum, minimum int32
	minimum = int32((^uint32(0)) >> 1)
	maximum = -minimum - 1
	var average, sum float32
	// receive the requests
	for {
		num, err := stream.Recv()
		if err == io.EOF {
			average = sum / float32(count)
			return stream.SendAndClose(&pb.StatisticsResponse{
				Count:   count,
				Maximum: maximum,
				Minimum: minimum,
				Average: average,
			})
		}
		if err != nil {
			log.Fatalf("failed to recv: %v", err)
			return err
		}
		count++
		if maximum < num.Number {
			maximum = num.Number
		}
		if minimum > num.Number {
			minimum = num.Number
		}
		sum += float32(num.Number)
	}
}

func (sms *SimpleMathServer) PrimeFactorization(stream pb.SimpleMath_PrimeFactorizationServer) error {
	// sending metadata to client: create trailer, using defer to record timestamp of function return
	defer func() {
		trailer := metadata.Pairs("timestamp", time.Now().Format(timestampFormat))
		stream.SetTrailer(trailer)
	}()

	// receiving metadata from client: read metadata from client
	md, _ := metadata.FromIncomingContext(stream.Context())
	if t, ok := md["timestamp"]; ok {
		log.Printf("timestamp from metadata:")
		for i, e := range t {
			log.Printf(" %d. %s", i, e)
		}
	}
	// sending metadata to client: create and send header
	header := metadata.New(map[string]string{"timestamp": time.Now().Format(timestampFormat)})
	stream.SendHeader(header)
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("failed to recv: %v", err)
			return err
		}
		stream.Send(&pb.PrimeFactorizationResponse{Result: primeFactorization(int(in.Number))})
	}
	return nil
}

func primeFactorization(num int) string {
	if num <= 2 {
		return strconv.Itoa(num)
	}
	n := num
	prefix := ""
	result := ""
	for i := 2; i <= n; i++ {
		for n != i {
			if n%i == 0 {
				result += prefix + strconv.Itoa(i)
				prefix = " * "
				n /= i
			} else {
				break
			}
		}
	}
	if result == "" {
		result = "1"
	}
	result = " = " + result + " * " + strconv.Itoa(n)
	return strconv.Itoa(num) + result
}
