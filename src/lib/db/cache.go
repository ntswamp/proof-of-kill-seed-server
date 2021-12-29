package lib_db

import (
	"fmt"
	"reflect"
)

/**
 * DBから取得したモデルを保管しておくもの.
 */
type ModelCacheStore struct {
	info  *ModelInfo
	store map[string]interface{}
}

func NewModelCacheStore(info *ModelInfo) *ModelCacheStore {
	return &ModelCacheStore{
		info:  info,
		store: map[string]interface{}{},
	}
}

/**
 * モデル1つをストアに保存.
 */
func (self *ModelCacheStore) saveModel(model interface{}) {
	pKeyLen := self.info.KeyLen()
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = reflect.Indirect(modelValue)
	} else {
		// storeに保存するのはポインタアドレス.
		tmp := reflect.New(modelValue.Type())
		tmp.Elem().Set(modelValue)
		model = tmp.Interface()
	}
	mapAddr := &self.store
	for i, field := range self.info.PrimaryFields {
		modelField := modelValue.FieldByName(field.Name)
		key := fmt.Sprintf("%v", modelField.Interface())
		if i == (pKeyLen - 1) {
			(*mapAddr)[key] = model
		} else {
			if _, exists := (*mapAddr)[key]; !exists {
				(*mapAddr)[key] = &map[string]interface{}{}
			}
			mapAddr = (*mapAddr)[key].(*map[string]interface{})
		}
	}
}

/**
 * モデルをストアに保存.
 */
func (self *ModelCacheStore) Save(model interface{}) {
	modelValue := reflect.ValueOf(model)
	switch modelValue.Kind() {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		// 複数.
		for i := 0; i < modelValue.Len(); i++ {
			self.saveModel(modelValue.Index(i).Interface())
		}
	default:
		// 単体.
		self.saveModel(model)
	}
}

/**
 * モデルをストアから取得.
 */
func (self *ModelCacheStore) get(pkey interface{}) []interface{} {
	pkeyValue := reflect.ValueOf(pkey)
	var keys []string
	if pkeyValue.Kind() == reflect.Array {
		// 複数取得.
		keys = make([]string, pkeyValue.Len())
		for i, _ := range keys {
			keys[i] = fmt.Sprintf("%v", pkeyValue.Index(i).Interface())
		}
	} else {
		// 単体取得.
		keys = []string{fmt.Sprintf("%v", pkey)}
	}
	mapAddr := &self.store
	for _, key := range keys {
		v, ok := (*mapAddr)[key]
		if !ok {
			// 存在しない.
			return []interface{}{}
		}
		switch v.(type) {
		case *map[string]interface{}:
			mapAddr = (*mapAddr)[key].(*map[string]interface{})
		default:
			// 探索終わり.
			return []interface{}{v}
		}
	}
	return aggregateStoreMapItems(mapAddr)
}

/**
 * モデルをストアから1件取得.
 */
func (self *ModelCacheStore) GetModel(pkey interface{}) interface{} {
	v := reflect.ValueOf(self.get(pkey))
	if v.Len() == 0 {
		return nil
	}
	return v.Index(0).Interface()
}

/**
 * モデルをストアから複数取得.
 */
func (self *ModelCacheStore) GetModels(pkeys interface{}) []interface{} {
	pkeyValues := reflect.ValueOf(pkeys)
	dest := []interface{}{}
	for i := 0; i < pkeyValues.Len(); i++ {
		arr := self.get(pkeyValues.Index(i).Interface())
		if 0 < len(arr) {
			dest = append(dest, arr...)
		}
	}
	return dest
}

/**
 * storeのmapの中身を集める.
 */
func aggregateStoreMapItems(mapAddr *map[string]interface{}) []interface{} {
	dest := []interface{}{}
	for _, v := range *mapAddr {
		switch v.(type) {
		case *map[string]interface{}:
			arr := aggregateStoreMapItems(v.(*map[string]interface{}))
			if 0 < len(arr) {
				dest = append(dest, arr...)
			}
		default:
			dest = append(dest, v)
		}
	}
	return dest
}
