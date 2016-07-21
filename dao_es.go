package tgo

import (
	"net/http"
	"sync"
	"time"

	"gopkg.in/olivere/elastic.v3"
)

type DaoES struct {
	IndexName string
	TypeName  string
}

var (
	esTransport *http.Transport
	esPoolMux   sync.Mutex
)

func (dao *DaoES) GetConnect() (*elastic.Client, error) {

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
	return client, err
}

func (dao *DaoES) Insert(id string, data interface{}) error {
	client, err := dao.GetConnect()

	if err != nil {
		return err
	}

	_, errRes := client.Index().Index(dao.IndexName).Type(dao.TypeName).Id(id).BodyJson(data).Do()

	if errRes != nil {
		UtilLogErrorf("insert error :%s", errRes.Error())
		return errRes
	}

	return nil
}
