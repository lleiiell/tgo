package tgo

import (
	"net/http"
	"sync"
	"time"

	"github.com/jolestar/go-commons-pool"
	"gopkg.in/olivere/elastic.v3"
)

type DaoES struct {
	IndexName string
	TypeName  string
}

var (
	esPool      *pool.ObjectPool
	esPoolMux   sync.Mutex
	esClient    *elastic.Client
	esClientMux sync.Mutex
)

func getESPoolConfig() *pool.ObjectPoolConfig {
	config := configESGet()

	return &pool.ObjectPoolConfig{
		Lifo:               config.ClientLifo,
		BlockWhenExhausted: true,
		MaxWaitMillis:      config.ClientMaxWaitMillis,
		MaxIdle:            config.ClientMaxIdle,
		MaxTotal:           config.ClientMaxTotal,
		TestOnBorrow:       true,
		MinIdle:            config.ClientMinIdle}
}

func (dao *DaoES) GetConnect() (*elastic.Client, error) {

	/*
		if esPool == nil {
			esPoolMux.Lock()

			defer esPoolMux.Unlock()

			if esPool == nil {
				configPool := getESPoolConfig()
				factory := &DaoESFactory{}
				esPool = pool.NewObjectPool(factory, configPool)
			}
		}

		//UtilLogInfof("active num :%d, idle num :%d", esPool.GetNumActive(), esPool.GetNumIdle())

		client, err := esPool.BorrowObject()

		if err != nil {
			UtilLogErrorf("get es client from pool failed:%s", err.Error())
			return nil, err
		}

		if client == nil {
			UtilLogErrorf("get es client from pool failed: client is nil")
			return nil, nil
		}

		esClient, ok := client.(*elastic.Client)

		if !ok {
			UtilLogErrorf("get es client from pool failed: convert client failed")

			return nil, errors.New("get client failed")
		}
		return esClient, nil*/

	config := configESGet()

	if esClient == nil {
		esClientMux.Lock()
		defer esClientMux.Unlock()

		if esClient == nil {
			clientHttp := &http.Client{
				Transport: &http.Transport{
					MaxIdleConnsPerHost: config.TransportMaxIdel,
				},
				Timeout: time.Duration(config.Timeout) * time.Millisecond,
			}
			client, err := elastic.NewClient(elastic.SetHttpClient(clientHttp), elastic.SetURL(config.Address...))
			if err != nil {
				// Handle error

				UtilLogErrorf("es connect error :%s,address:%v", err.Error(), config)

				return nil, err
			}
			esClient = client
		}

	}

	return esClient, nil
}

func (dao *DaoES) CloseConnect(client *elastic.Client) {
	//esPool.ReturnObject(client)
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

func (dao *DaoES) Update(id string, doc interface{}) error {
	client, err := dao.GetConnect()

	if err != nil {
		return err
	}
	defer dao.CloseConnect(client)
	_, errRes := client.Update().Index(dao.IndexName).Type(dao.TypeName).Id(id).
		Doc(doc).
		Do()

	if errRes != nil {
		UtilLogErrorf("daoES Update error :%s", errRes.Error())
		return err
	}

	return nil
}

func (dao *DaoES) UpdateAppend(id string, name string, value interface{}) error {
	client, err := dao.GetConnect()

	if err != nil {
		return err
	}

	_, errRes := client.Update().Index(dao.IndexName).Type(dao.TypeName).Id(id).
		Script(elastic.NewScriptFile("append-reply").Param("reply", value)).
		Do()

	if errRes != nil {
		UtilLogErrorf("daoES Update error :%s", errRes.Error())
		return err
	}

	return nil
}
