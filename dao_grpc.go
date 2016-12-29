package tgo

import (
	"google.golang.org/grpc"

	"github.com/jolestar/go-commons-pool"

	"errors"

	"sync"
)

type DaoGRPC struct {
	ServerName  string
	Config      *ConfigPool
	DialOptions []grpc.DialOption
}

var (
	grpcPoolMap map[string]*pool.ObjectPool
	grpcPoolMux sync.Mutex
)

func daoGRPCGetConfig(serverName string) (*ConfigPool, error) {

	poolName := "grpc-" + serverName

	config := configPoolGet(poolName)

	if config == nil {
		return nil, errors.New("pool config is null: " + poolName)
	}
	return config, nil
}

func (dao *DaoGRPC) GetConn() (*grpc.ClientConn, error) {

	if grpcPoolMap == nil {
		grpcPoolMux.Lock()
		if grpcPoolMap == nil {
			grpcPoolMap = make(map[string]*pool.ObjectPool)
		}
		grpcPoolMux.Unlock()
	}

	grpcPool, ok := grpcPoolMap[dao.ServerName]

	if !ok || grpcPool == nil {
		grpcPoolMux.Lock()

		defer grpcPoolMux.Unlock()

		grpcPool, ok = grpcPoolMap[dao.ServerName]

		if !ok || grpcPool == nil {

			config, err := daoGRPCGetConfig(dao.ServerName)
			if err != nil {
				return nil, err
			}

			pc := &pool.ObjectPoolConfig{
				Lifo:               config.Lifo,
				BlockWhenExhausted: config.BlockWhenExhausted,
				MaxWaitMillis:      config.MaxWaitMillis,
				MaxIdle:            config.MaxIdle,
				MaxTotal:           config.MaxTotal,
				TestOnBorrow:       config.TestOnBorrow,
				TestOnCreate:       config.TestOnCreate,
				TestOnReturn:       config.TestOnReturn,
				MinIdle:            config.MinIdle}

			factory := &DaoGRPCFactory{}
			factory.Config = config
			factory.DialOptions = dao.DialOptions
			grpcPoolMap[dao.ServerName] = pool.NewObjectPool(factory, pc)

			grpcPool, ok = grpcPoolMap[dao.ServerName]

			if !ok || grpcPool == nil {
				return nil, errors.New("create grpc pool failed:" + dao.ServerName)
			}

		}
	}
	conn, err := grpcPool.BorrowObject()

	if err != nil {
		UtilLogErrorf("get grpc conn from pool  failed, server name :%s ,err:%s", dao.ServerName, err.Error())
		return nil, err
	}
	//http2 共用同一链接，所以拿到即可还回
	//grpcPool.ReturnObject(conn)

	if conn == nil {
		UtilLogErrorf("get grpc conn from pool failed: conn is nil")
		return nil, nil
	}

	grpcConn, ok := conn.(*grpc.ClientConn)

	if !ok {
		errMsg := "get grpc conn from pool failed: convert client failed"
		UtilLogError(errMsg)

		return nil, errors.New(errMsg)
	}

	return grpcConn, nil

}

func (dao *DaoGRPC) CloseConn(conn *grpc.ClientConn) error {
	grpcPool, ok := grpcPoolMap[dao.ServerName]

	if !ok {
		return errors.New("grpc pool is not exist")
	}
	return grpcPool.ReturnObject(conn)
}
