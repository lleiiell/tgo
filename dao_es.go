package tgo

import (
	"sync"
	"errors"
	"gopkg.in/olivere/elastic.v3"
	"github.com/jolestar/go-commons-pool"
)

type DaoES struct {
	IndexName string
	TypeName  string
}

var (
	esPool *pool.ObjectPool
	esPoolMux   sync.Mutex
)

func  getESPoolConfig() *pool.ObjectPoolConfig{
	config := configESGet()

	return &pool.ObjectPoolConfig{
		Lifo:config.ClientLifo,
		MaxIdle:config.ClientMaxIdle,
		MaxTotal:config.ClientMaxTotal,
		MinIdle:config.ClientMinIdle}
}

func (dao *DaoES) GetConnect() (*elastic.Client, error) {

	if esPool ==nil{
		esPoolMux.Lock()

		defer esPoolMux.Unlock()

		if esPool ==nil{
			configPool := getESPoolConfig()
			factory:=&DaoESFactory{}
			esPool = pool.NewObjectPool(factory, configPool)
		}
	}

	client,err:=esPool.BorrowObject()

	if err!=nil{
		UtilLogErrorf("get es client from pool failed:%s", err.Error())
		return nil,err
	}

	if client ==nil{
		UtilLogErrorf("get es client from pool failed: client is nil")
		return nil,nil
	}

	esClient,ok := client.(*elastic.Client)

	if !ok{
		UtilLogErrorf("get es client from pool failed: convert client failed")

		return nil,errors.New("get client failed")
	}
	return esClient,nil
	/*
	address := configESGetAddress()

	if esTransport == nil {
		esPoolMux.Lock()
		defer esPoolMux.Unlock()
		if esTransport == nil {
			esTransport = &http.Transport{
				MaxIdleConnsPerHost: 50,
			}
		}
	}
	clientHttp := &http.Client{
		Transport: esTransport,
		Timeout:   time.Duration(1000) * time.Millisecond,
	}

	client, err := elastic.NewClient(elastic.SetHttpClient(clientHttp), elastic.SetSniff(false), elastic.SetURL(address...))

	if err != nil {
		// Handle error

		UtilLogErrorf("es connect error :%s,address:%v", err.Error(), address)

		return nil, err
	}
	return client, err */
}

func (dao *DaoES) CloseConnect(client *elastic.Client){
	esPool.ReturnObject(client)
}

func (dao *DaoES) Insert(id string, data interface{}) error {
	client, err := dao.GetConnect()

	if err != nil {
		return err
	}
	defer dao.CloseConnect(client)
	
	_, errRes := client.Index().Index(dao.IndexName).Type(dao.TypeName).Id(id).BodyJson(data).Do()

	if errRes != nil {
		UtilLogErrorf("insert error :%s", errRes.Error())
		return errRes
	}

	return nil
}
