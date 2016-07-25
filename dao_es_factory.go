package tgo

import(
"github.com/jolestar/go-commons-pool"
"gopkg.in/olivere/elastic.v3"
"net/http"
"time"
)


type DaoESFactory struct{

}

func (f *DaoESFactory) MakeObject() (*pool.PooledObject, error) {
  client,err:= f.MakeClient()
	return pool.NewPooledObject(client), err
}

func (f *DaoESFactory) MakeClient()(*elastic.Client,error){
  config:= configESGet()

	clientHttp := &http.Client{
		Transport: &http.Transport{
            MaxIdleConnsPerHost: config.TransportMaxIdel,
    	},
		Timeout:   time.Duration(config.Timeout) * time.Millisecond,
	}

	client, err := elastic.NewClient(elastic.SetHttpClient(clientHttp), elastic.SetSniff(false), elastic.SetURL(config.Address...))

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
