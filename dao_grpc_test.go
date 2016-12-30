package tgo

import (
	/*"context"
	"fmt"*/
	"google.golang.org/grpc"
	/*pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"log"
	"sync"*/
	"log"
	"testing"
)

func TestDaoGRPC_GetConn(t *testing.T) {
	daoGrpc := &DaoGRPC{}
	daoGrpc.DialOptions = append(daoGrpc.DialOptions, grpc.WithInsecure())
	daoGrpc.ServerName = "test"

	conn, err := daoGrpc.GetConn()

	if err != nil {
		t.Errorf("get failed:%s", err.Error())
	} else {
		log.Printf("conn:%v\n", conn)
		defer daoGrpc.CloseConn(conn)
	}
}
func BenchmarkDaoGRPC_GetConn(b *testing.B) {
	daoGrpc := &DaoGRPC{}
	daoGrpc.DialOptions = append(daoGrpc.DialOptions, grpc.WithInsecure())
	daoGrpc.ServerName = "test"

	for i := 0; i < b.N; i++ {
		conn, err := daoGrpc.GetConn()

		if err != nil {
			b.Errorf("get failed:%s", err.Error())
		} else {
			log.Printf("conn:%v\n", conn)
			daoGrpc.CloseConn(conn)
		}
	}

}

/*
func TestDaoGRPC_HelloWorld(t *testing.T) {
	daoGrpc := &DaoGRPC{}
	daoGrpc.DialOptions = append(daoGrpc.DialOptions, grpc.WithInsecure())
	daoGrpc.ServerName = "test"

	conn, err := daoGrpc.GetConn()

	if err != nil {
		t.Errorf("get conn failed:%s", err.Error())
	}
	defer daoGrpc.CloseConn(conn)

	c := pb.NewGreeterClient(conn)

	var reply *pb.HelloReply
	reply, err = c.SayHello(context.Background(), &pb.HelloRequest{Name: "tony"})

	if err != nil {
		t.Errorf("call hello failed :%s", err.Error())
	}
	log.Printf("result is %s \n", reply.Message)
}

func BenchmarkDaoGRPC_HelloWorld(b *testing.B) {
	var wg sync.WaitGroup

	for j := 0; j < 1000; j++ {
		wg.Add(1)
		f1 := func(index int) {
			defer wg.Done()
			for i := 0; i < b.N/1000; i++ {

				daoGrpc := &DaoGRPC{}
				daoGrpc.DialOptions = append(daoGrpc.DialOptions, grpc.WithInsecure())
				daoGrpc.ServerName = "test"

				conn, err := daoGrpc.GetConn()

				if err != nil {
					b.Errorf("get conn failed:%s，index:%d-%d", err.Error(), index, i)
					return
				}

				c := pb.NewGreeterClient(conn)

				//var reply *pb.HelloReply
				_, err = c.SayHello(context.Background(), &pb.HelloRequest{Name: fmt.Sprintf("tony-%d-%d", index, i)})

				if err != nil {
					b.Errorf("call hello failed :%s-%d-%d", err.Error(), index, i)
				}

				daoGrpc.CloseConn(conn)

			}
		}
		go f1(j)

	}
	wg.Wait()
}
*/
