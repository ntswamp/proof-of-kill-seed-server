package lib_db_test

import (
	"app/src/constant"
	lib_db "app/src/lib/db"
	lib_redis "app/src/lib/redis"
	"app/src/model"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DummyMultiKeyTable struct {
	model.BaseModel
	Id   uint32 `gorm:"primary_key:test;type:int unsigned"`
	Id2  uint32 `gorm:"primary_key:test;type:int unsigned"`
	Hoge string
}
type DummyMasterTable struct {
	model.BaseMaster
	Id   uint32 `gorm:"primary_key"`
	Hoge string
}
type DummyTable struct {
	Id   uint32 `gorm:"primary_key"`
	Id2  uint32
	Hoge string
}
type DummyTable2 struct {
	model.BaseModel
	Id  uint32        `gorm:"primary_key"`
	Loc *lib_db.Point `gorm:"type:point"`
}

// ModelCacheStoreのテスト.
func TestModelCacheStore(t *testing.T) {
	// DBに接続.
	client, err := lib_db.Connect(constant.DbIdolverse, nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer client.Close()

	db := client.GetDB()
	// テーブルを作成.
	db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&DummyTable{})
	defer db.DropTable(&DummyTable{})

	model := DummyTable{
		Id:   1000,
		Hoge: "テストです",
	}
	info := lib_db.NewModelInfo(db, model)

	store := info.CacheStore()
	store.Save(&model)

	cachedModel := store.GetModel(model.Id)
	assert.NotNil(t, cachedModel)

	model2, _ := cachedModel.(*DummyTable)
	assert.NotNil(t, model2)

	model.Hoge = "hogehogehoge"
	assert.Equal(t, model2.Hoge, "hogehogehoge")
}

// ModelManagerのテスト.
func TestModelManager(t *testing.T) {
	// redis.
	redisCache, err := lib_redis.NewClient(constant.RedisCache)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer redisCache.Terminate()
	// DBに接続.
	client, err := lib_db.Connect(constant.DbIdolverse, redisCache)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer client.Close()

	db := client.GetDB()
	// テーブルを作成.
	db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&DummyMultiKeyTable{})
	defer db.DropTable(&DummyMultiKeyTable{})

	manager := client.GetModelManager()
	// 複合プライマリキーのレコードを1件作成.
	record := DummyMultiKeyTable{
		Id:   1000,
		Id2:  1000,
		Hoge: "aiueo",
	}
	err = manager.CachedSave(&record, nil)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	err = manager.WriteAll()
	if err != nil {
		t.Fatalf(err.Error())
		return
	}

	// redisから取得.
	records := []*DummyMultiKeyTable{}
	err = manager.GetModels(&records, [][]uint32{
		[]uint32{record.Id, record.Id2},
	})
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	assert.Equal(t, len(records), 1)
	// DBから取得.
	records = []*DummyMultiKeyTable{}
	err = manager.GetModels(&records, []uint32{record.Id})
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	assert.Equal(t, len(records), 1)
	// ModelManagerのキャッシュから取得.
	records2 := []*DummyMultiKeyTable{}
	err = manager.GetModels(&records2, []uint32{record.Id})
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	assert.Equal(t, len(records2), 1)
	if len(records) == 0 || len(records2) == 0 {
		return
	}
	// 同じインスタンスのアドレスのはず.
	assert.Equal(t, reflect.ValueOf(records[0]).Pointer(), reflect.ValueOf(records2[0]).Pointer())
	// 単体取得.
	record2 := &DummyMultiKeyTable{}
	ins, err := manager.GetModel(record2, record.Id)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	assert.NotNil(t, ins)
	if ins == nil {
		return
	}
	assert.Equal(t, record2.Id, record.Id)
	// 単体取得.存在しない場合.
	record2 = &DummyMultiKeyTable{}
	ins, err = manager.GetModel(record2, record.Id+1)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	assert.Nil(t, ins)

	// 削除.
	err = manager.SetDelete(&record)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	err = manager.WriteAll()
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	// DBから取得.
	records = []*DummyMultiKeyTable{}
	err = manager.GetModels(&records, []uint32{record.Id})
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	assert.Equal(t, len(records), 0)
}

// ModelManagerのテスト(マスターデータ).
func TestModelManagerForMaster(t *testing.T) {

	// DBに接続.
	client, err := lib_db.Connect(constant.DbIdolverse, nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer client.Close()

	db := client.GetDB()
	// テーブルを作成.
	db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&DummyMasterTable{})
	defer db.DropTable(&DummyMasterTable{})

	// マスターデータのレコードを1件作成.
	master := &DummyMasterTable{
		Id:   1,
		Hoge: "aiueo",
	}
	err = db.Save(master).Error
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	recordId := master.Id

	manager := lib_db.NewModelManager(db, nil, true)
	// DBから取得.
	records := []*DummyMasterTable{}
	err = manager.GetModels(&records, []uint32{recordId})
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	assert.Equal(t, len(records), 1)
	// キャッシュから取得.
	records2 := []*DummyMasterTable{}
	err = manager.GetModels(&records2, []uint32{recordId})
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	assert.Equal(t, len(records2), 1)
	if len(records) == 0 || len(records2) == 0 {
		return
	}
	// 同じインスタンスのアドレスのはず.
	assert.Equal(t, reflect.ValueOf(records[0]).Pointer(), reflect.ValueOf(records2[0]).Pointer())
	// localcacheから取得.
	manager = lib_db.NewModelManager(db, nil, true)
	records2 = []*DummyMasterTable{}
	err = manager.GetModels(&records2, []uint32{recordId})
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	assert.Equal(t, len(records2), 1)
}

// lib_db_fields.PointのSelectのテスト.
func TestCustomTypePoint(t *testing.T) {
	// DBに接続.
	client, err := lib_db.Connect(constant.DbIdolverse, nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer client.Close()

	db := client.GetDB()
	// テーブルを作成.
	db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&DummyTable2{})
	defer db.DropTable(&DummyTable2{})
	manager := client.GetModelManager()
	// テスト用にレコードをinsert.
	record := &DummyTable2{}
	record.Id = 1
	record.Loc = &lib_db.Point{
		Lat: 35.028611,
		Lon: 139.008056,
	}
	err = manager.CachedSave(record, lib_db.OptInsert)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = manager.WriteAll()
	if err != nil {
		t.Fatal(err)
		return
	}
	// 更新.
	record.Loc.Lat = 35.628611
	record.Loc.Lon = 139.708056
	err = manager.CachedSave(record, nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = manager.WriteAll()
	if err != nil {
		t.Fatal(err)
		return
	}
	// DBから取得する.
	manager = lib_db.NewModelManager(db, nil, true)
	ins, err := manager.GetModel(&DummyTable2{}, 1)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}
	assert.NotNil(t, ins)
	if ins == nil {
		return
	}
	record = ins.(*DummyTable2)
	assert.NotNil(t, record.Loc)
	if record.Loc == nil {
		return
	}
	assert.Equal(t, record.Loc.Lat, float64(35.628611))
	assert.Equal(t, record.Loc.Lon, float64(139.708056))
}
