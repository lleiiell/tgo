package tgo

import (
  "gopkg.in/olivere/elastic.v3"
)

type DaoES struct {
	IndexName string
  TypeName string
}

func (dao *DaoES) GetConnect()(*elastic.Client, error){

  address:=configESGetAddress()

	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(address...))

	if err != nil {
		// Handle error

		UtilLogErrorf("es connect error :%s,address:%v", err.Error(),address)

		return nil, err
	}
  return client,err
}

func (dao *DaoES) Insert(id string,data interface{}) error{
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
