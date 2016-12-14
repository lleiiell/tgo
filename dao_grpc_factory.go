package tgo

import (
	"github.com/jolestar/go-commons-pool"

	"errors"
	"google.golang.org/grpc"
)

type DaoGRPCFactory struct {
	Config      *ConfigPool
	DialOptions []grpc.DialOption
}

func (f *DaoGRPCFactory) MakeObject() (*pool.PooledObject, error) {
	client, err := f.MakeConn()
	if err != nil {
		return nil, err
	}
	return pool.NewPooledObject(client), err
}

func (f *DaoGRPCFactory) MakeConn() (*grpc.ClientConn, error) {
	address, errAddress := f.Config.GetAddressRandom()

	if errAddress != nil {
		return nil, errAddress
	}

	conn, err := grpc.Dial(address, f.DialOptions...)

	if err != nil {
		// Handle error

		return nil, err
	}
	return conn, err
}

func (f *DaoGRPCFactory) DestroyObject(object *pool.PooledObject) error {
	//do destroy
	conn, ok := object.Object.(*grpc.ClientConn)

	if !ok {
		errMsg := "grpc pool destory object failed,convert to clientConn failed"
		UtilLogInfo(errMsg)
		return errors.New(errMsg)
	}
	return conn.Close()
}

func (f *DaoGRPCFactory) ValidateObject(object *pool.PooledObject) bool {
	//do validate

	return true
}

func (f *DaoGRPCFactory) ActivateObject(object *pool.PooledObject) error {
	//do activate
	return nil
}

func (f *DaoGRPCFactory) PassivateObject(object *pool.PooledObject) error {
	//do passivate
	return nil
}
