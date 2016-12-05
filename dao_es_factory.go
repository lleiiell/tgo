package tgo

import (
	"net/http"
	"sync"
	"time"

	"github.com/jolestar/go-commons-pool"
	"gopkg.in/olivere/elastic.v3"
)

var (
	esTransport    *http.Transport
	esTransportMux sync.Mutex
)

type DaoESFactory struct {
}

func (f *DaoESFactory) MakeObject() (*pool.PooledObject, error) {
	client, err := f.MakeClient()
	return pool.NewPooledObject(client), err
}

func (f *DaoESFactory) MakeClient() (*elastic.Client, error) {
	config := configESGet()

	if esTransport == nil {
		esTransportMux.Lock()

		defer esTransportMux.Unlock()

		if esTransport == nil {
			esTransport = &http.Transport{
				MaxIdleConnsPerHost: config.TransportMaxIdel,
			}
		}
	}
	clientHttp := &http.Client{
		Transport: esTransport,
		Timeout:   time.Duration(config.Timeout) * time.Millisecond,
	}

	client, err := elastic.NewClient(elastic.SetHttpClient(clientHttp), elastic.SetURL(config.Address...))

	if err != nil {
		// Handle error

		UtilLogErrorf("es connect error :%s,address:%v", err.Error(), config.Address)

		return nil, err
	}
	return client, err
}

func (f *DaoESFactory) DestroyObject(object *pool.PooledObject) error {
	//do destroy

	return nil
}

func (f *DaoESFactory) ValidateObject(object *pool.PooledObject) bool {
	//do validate
	esClient, ok := object.Object.(*elastic.Client)

	if !ok {
		UtilLogInfo("es pool validate object failed,convert to client failed")
		return false
	}
	if !esClient.IsRunning() {
		UtilLogInfo("es pool validate object failed,convert to client failed")
		return false
	}

	return true
}

func (f *DaoESFactory) ActivateObject(object *pool.PooledObject) error {
	//do activate
	return nil
}

func (f *DaoESFactory) PassivateObject(object *pool.PooledObject) error {
	//do passivate
	return nil
}
