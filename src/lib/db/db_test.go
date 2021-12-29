package lib_db

import (
	"app/src/constant"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DummyTable struct {
	Id   uint32 `gorm:"primary_key"`
	Id2  uint32
	Hoge string
}

func TestDatabase(t *testing.T) {
	// DBに接続.
	client, err := Connect(constant.DbIdolverse, nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer client.Close()

	db := client.GetDB()
	// テーブルを作成.
	db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&DummyTable{})
	defer db.DropTable(&DummyTable{})

	// insert.
	record := &DummyTable{
		Id:   12345,
		Id2:  12345678,
		Hoge: "hogehoge",
	}
	err = db.Create(record).Error
	if err != nil {
		t.Fatal(err)
		return
	}

	// select.
	record = &DummyTable{}
	err = db.Find(record, 12345).Error
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Equal(t, record.Hoge, "hogehoge")

	record = &DummyTable{}
	err = db.Where("(`id`,`id2`) in ((?),(?))", []interface{}{[]uint32{12345, 12345678}, []uint32{12346, 12345679}}...).Find(record).Error
	if err != nil {
		t.Fatal(err)
		return
	}
	assert.Equal(t, record.Hoge, "hogehoge")

}

func TestStartTransaction(t *testing.T) {
	client, err := Connect(constant.DbIdolverse, nil)
	if err != nil {
		t.Log(err)
	}

	client.StartTransaction()

	modelMgr := client.GetModelManager()

	err = modelMgr.WriteAll()
	if err != nil {
		t.Log(err)
	}
	err = client.CommitTransaction()
	if err != nil {
		t.Log(err)
	}

	err = client.RollbackTransaction()
	if err != nil {
		t.Log(err)
	}
	client.Close()
}
