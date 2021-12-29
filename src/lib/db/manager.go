package lib_db

import (
	"app/src/constant"
	lib_error "app/src/lib/error"
	lib_redis "app/src/lib/redis"
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
)

const DefaultCacheTTL = 86400

type ModelManager struct {
	db                    *gorm.DB
	redis                 *lib_redis.Client
	modelInfos            map[string]*ModelInfo
	reservedSaveModelKeys []string
	reservedSaveInfos     map[string]*ReservedInfo
	reservedDelModelKeys  []string
	reservedDelModels     map[string]interface{}
	insertModelCount      uint32
	writeEndTasks         []*Task
	CacheTTL              int64
	writable              bool
}

/**
 * ModelManagerを作成.
 * redisがnilのときはモデル(マスター以外)のキャッシュは行いません.
 */
func NewModelManager(db *gorm.DB, redis *lib_redis.Client, writable bool) *ModelManager {
	return &ModelManager{
		db:                    db,
		redis:                 redis,
		modelInfos:            map[string]*ModelInfo{},
		reservedSaveModelKeys: []string{},
		reservedSaveInfos:     map[string]*ReservedInfo{},
		reservedDelModelKeys:  []string{},
		reservedDelModels:     map[string]interface{}{},
		insertModelCount:      0,
		CacheTTL:              DefaultCacheTTL,
		writable:              writable,
	}
}

func (self *ModelManager) log(format string, args ...interface{}) {
	if !constant.IsLocal() {
		return
	}
	fmt.Println(fmt.Sprintf(format, args...))
}

/**
 * DBのコネクションを取得.
 */
func (self *ModelManager) GetDB() *gorm.DB {
	return self.db
}

/**
 * キャッシュ用redisクライアントを取得.
 */
func (self *ModelManager) GetRedis() *lib_redis.Client {
	return self.redis
}

/**
 * モデル情報を取得.
 */
func (self *ModelManager) getInfo(model interface{}) *ModelInfo {
	name := self.db.NewScope(model).QuotedTableName()
	if _, ok := self.modelInfos[name]; !ok {
		self.modelInfos[name] = NewModelInfo(self.db, model)
	}
	return self.modelInfos[name]
}

/**
 * テーブル名を取得.
 */
func (self *ModelManager) GetTableName(model interface{}) string {
	return self.getInfo(model).TableName
}

/**
 * モデル名を取得.
 */
func (self *ModelManager) GetModelName(model interface{}) string {
	v := reflect.ValueOf(model)
	t := v.Type()
	for t.Kind() != reflect.Struct {
		t = t.Elem()
	}
	return t.Name()
}

/**
 * すべてのカラムを取得するselect句を付ける.
 */
func (self *ModelManager) Select(db *gorm.DB, model interface{}) *gorm.DB {
	info := self.getInfo(model)
	return info.Select(db)
}

/**
 * 溜め込んだモデルの識別子.
 */
func (self *ModelManager) makeKey(model interface{}) string {
	info := self.getInfo(model)
	var pkey string
	if self.db.NewScope(model).PrimaryKeyZero() {
		pkey = fmt.Sprintf("NewModel:%v", self.insertModelCount)
		self.insertModelCount++
	} else {
		pkey = info.MakePrimaryKeyStringByModel(model, 0)
	}
	return fmt.Sprintf("%s:%s", info.TableName, pkey)
}

/**
 * プライマリキーを指定して取得.
 * 存在の有無を返す.
 */
func (self *ModelManager) GetModel(dest interface{}, pkey interface{}) (interface{}, error) {
	destV := reflect.ValueOf(dest)
	destSlicePtr := reflect.New(reflect.SliceOf(destV.Type()))
	keySliceValue := reflect.MakeSlice(reflect.SliceOf(reflect.ValueOf(pkey).Type()), 1, 1)
	keySliceValue.Index(0).Set(reflect.ValueOf(pkey))
	err := self.GetModels(destSlicePtr.Interface(), keySliceValue.Interface())
	if err != nil {
		return false, err
	}
	elem := destV.Elem()
	if destSlicePtr.Elem().Len() == 0 {
		// 無かった.
		return nil, nil
	} else {
		result := destSlicePtr.Elem().Index(0)
		elem.Set(result.Elem())
		return result.Interface(), nil
	}
}

/**
 * プライマリキーを指定してまとめて取得.
 * 複合プライマリキーのテーブルで,プライマリキーの数よりも少ないプライマリキーを指定した場合は,ほしい値が取れない場合があるので注意.
 */
func (self *ModelManager) GetModels(dest interface{}, pkeys interface{}) error {
	destV := reflect.ValueOf(dest)
	elem := destV.Elem()
	sl := reflect.MakeSlice(elem.Type(), 0, 0)
	info := self.getInfo(dest)
	self.log(info.TableName)
	keyLen := info.KeyLen()
	// プライマリキーの有無確認用にmapを用意.
	keyValues := reflect.ValueOf(pkeys)
	keyCnt := keyValues.Len()
	keyStrings := make(map[string]bool, keyCnt)
	for i := 0; i < keyCnt; i++ {
		keyString := info.MakePrimaryKeyString(keyValues.Index(i).Interface())
		keyStrings[keyString] = false
	}
	// すでに取得済みのモデルを取得.
	cachedModels := info.CacheStore().GetModels(pkeys)
	for _, cachedModel := range cachedModels {
		for i := 0; i < keyLen; i++ {
			keyString := info.MakePrimaryKeyStringByModel(cachedModel, i+1)
			if _, ok := keyStrings[keyString]; ok {
				keyStrings[keyString] = true
			}
		}
		sl = reflect.Append(sl, reflect.ValueOf(cachedModel))
	}
	// 未取得のキーを集める.
	notExists := map[int][]interface{}{}
	for i := 0; i < keyCnt; i++ {
		keyValue := keyValues.Index(i)
		key := keyValue.Interface()
		keyString := info.MakePrimaryKeyString(key)
		if keyStrings[keyString] {
			// 取得済み.
			continue
		}
		length := 1
		if keyValue.Kind() == reflect.Array || keyValue.Kind() == reflect.Slice {
			length = keyValue.Len()
		}
		if _, ok := notExists[length]; !ok {
			notExists[length] = []interface{}{}
		}
		notExists[length] = append(notExists[length], key)
	}
	if 0 < len(notExists) {
		self.log("Get from cache...%v", len(notExists))
		if info.GetIsMaster(dest) {
			// マスターデータはlocalcacheから取得.
			mKeys, _ := notExists[keyLen]
			if 0 < len(mKeys) {
				cacheStrings := make([]string, len(mKeys))
				for i, mKey := range mKeys {
					cacheStrings[i] = fmt.Sprintf("%s##%s", info.TableName, info.MakePrimaryKeyString(mKey))
				}
				tmpSlicePtr := reflect.New(elem.Type()).Interface()
				lib_redis.GetLocalCacheMany(cacheStrings, tmpSlicePtr)
				tmpSlice := reflect.ValueOf(tmpSlicePtr).Elem()
				sl = reflect.AppendSlice(sl, tmpSlice)
				exists := make(map[string]bool, tmpSlice.Len())
				for i := 0; i < tmpSlice.Len(); i++ {
					exists[info.MakePrimaryKeyStringByModel(tmpSlice.Index(i).Interface(), 0)] = true
				}
				notExists[keyLen] = []interface{}{}
				for _, mKey := range mKeys {
					ks := info.MakePrimaryKeyString(mKey)
					if _, ok := exists[ks]; ok {
						// localcacheから取得できた.
						continue
					}
					notExists[keyLen] = append(notExists[keyLen], mKey)
				}
			}
		} else if self.redis != nil {
			// トランデータはredisから取得.
			mKeys, _ := notExists[keyLen]
			if 0 < len(mKeys) {
				self.log("from redis...%v", mKeys)
				cacheKeys := make([]interface{}, len(mKeys))
				for i, mKey := range mKeys {
					cacheKeys[i] = fmt.Sprintf("%s:%s", info.TableName, info.MakePrimaryKeyString(mKey))
				}
				tmpSlicePtr := reflect.New(elem.Type()).Interface()
				err := self.redis.MGet(tmpSlicePtr, cacheKeys...)
				if err != nil {
					return lib_error.WrapError(err)
				}
				tmpSlice := reflect.ValueOf(tmpSlicePtr).Elem()
				notExists[keyLen] = []interface{}{}
				for i := 0; i < tmpSlice.Len(); i++ {
					vElem := tmpSlice.Index(i)
					if vElem.IsNil() {
						// キャッシュになかった.
						notExists[keyLen] = append(notExists[keyLen], mKeys[i])
						self.log("not found:%v", mKeys[i])
						continue
					}
					sl = reflect.Append(sl, vElem)
				}
			}
		}
	}
	if 0 < len(notExists) {
		for length, keys := range notExists {
			if len(keys) == 0 {
				continue
			}
			self.log("keys:%v:%v", keys, length)
			tmpSlicePtr := reflect.New(elem.Type()).Interface()
			scope := info.Select(info.Query(self.db, keys, length)).Set("gorm:auto_preload", true).Find(tmpSlicePtr)
			if scope.Error != nil && scope.RecordNotFound() {
				return lib_error.WrapError(scope.Error)
			}
			tmpSlice := reflect.ValueOf(tmpSlicePtr).Elem()
			info.CacheStore().Save(tmpSlice.Interface())
			if info.GetIsMaster(dest) {
				// localcacheに保存.
				for i := 0; i < tmpSlice.Len(); i++ {
					data := tmpSlice.Index(i).Interface()
					k := fmt.Sprintf("%s##%s", info.TableName, info.MakePrimaryKeyStringByModel(data, 0))
					err := lib_redis.SetLocalCache(k, data)
					if err != nil {
						return lib_error.WrapError(err)
					}
				}
			} else if self.redis != nil && keyLen == length {
				// redisに保存.
				models := make([]interface{}, tmpSlice.Len())
				for i := 0; i < tmpSlice.Len(); i++ {
					models[i] = tmpSlice.Index(i).Interface()
				}
				err := self.SaveModelsToCache(models, self.CacheTTL, false)
				if err != nil {
					return lib_error.WrapError(err)
				}
			}
			sl = reflect.AppendSlice(sl, tmpSlice)
		}
	}
	elem.Set(sl)
	return nil
}

/**
 * テーブルのマスターデータをすべて取得.
 */
func (self *ModelManager) GetMasterModelAll(dest interface{}, reload bool) error {
	var err error = nil
	info := self.getInfo(dest)
	cacheKey := fmt.Sprintf("mastermodel_keyset::%s", info.TableName)
	var pkeys interface{} = nil
	if self.redis != nil && !reload {
		// キャッシュからIDを取得.
		exists, err := self.redis.Exists(cacheKey)
		if err != nil {
			return err
		} else if exists {
			tmp := []interface{}{}
			_, err = self.redis.Get(&tmp, cacheKey)
			if err != nil {
				return err
			}
			pkeys = tmp
		}
	}
	if pkeys != nil {
		// キャッシュあり.
		return self.GetModels(dest, pkeys)
	}
	// 取り直し.
	err = self.db.Find(dest).Error
	if err != nil || self.redis == nil {
		return err
	}
	// キャッシュに保存.
	destV := reflect.ValueOf(dest).Elem()
	keys := make([]interface{}, destV.Len())
	for i, _ := range keys {
		destModel := destV.Index(i)
		if destModel.Kind() == reflect.Ptr {
			destModel = reflect.Indirect(destModel)
		}
		if info.KeyLen() == 1 {
			keys[i] = destModel.FieldByName(info.PrimaryFields[0].Name).Interface()
		} else {
			destModelKeys := make([]interface{}, info.KeyLen())
			for j, field := range info.PrimaryFields {
				destModelKeys[j] = destModel.FieldByName(field.Name).Interface()
			}
			keys[i] = destModelKeys
		}
	}
	err = self.redis.Set(cacheKey, keys)
	if err == nil {
		err = self.redis.Expire(cacheKey, self.CacheTTL)
	}
	return err
}

/**
 * 保存予約情報の取得.
 */
func (self *ModelManager) getReservedInfo(model interface{}) (*ReservedInfo, error) {
	// 識別子.
	key := self.makeKey(model)
	// 既に登録されている予約情報.
	info, _ := self.reservedSaveInfos[key]
	if info != nil {
		info.Model = model
		return info, nil
	}
	// 削除予約されていないかを確認.
	d, _ := self.reservedDelModels[key]
	if d != nil {
		return nil, lib_error.NewAppError(lib_error.DefaultErrorCode, "削除予約に入ってる:%s", key)
	}
	// 新規作成.
	info = &ReservedInfo{
		Model:       model,
		fieldMap:    map[string]bool{},
		ForceInsert: false,
		ForceUpdate: false,
		SavedTasks:  []*Task{},
	}
	self.reservedSaveInfos[key] = info
	self.reservedSaveModelKeys = append(self.reservedSaveModelKeys, key)
	return info, nil
}

/**
 * 保存するモデルに追加.
 */
func (self *ModelManager) CachedSave(model interface{}, opt *SaveOptions) error {
	if opt == nil {
		opt = &SaveOptions{}
	} else if opt.ForceInsert && opt.ForceUpdate {
		return lib_error.NewAppError(lib_error.DefaultErrorCode, "ForceInsertとForceUpdateは両方をtrueに出来ません")
	}
	if opt.Fields == nil {
		opt.Fields = []string{}
	}
	info, err := self.getReservedInfo(model)
	if err != nil {
		return err
	}
	info.ForceInsert = info.ForceInsert || opt.ForceInsert
	info.ForceUpdate = info.ForceUpdate || opt.ForceUpdate
	if info.ForceInsert && info.ForceUpdate {
		return lib_error.NewAppError(lib_error.DefaultErrorCode, "ForceInsertとForceUpdateは両方をtrueで指定していないけど、複数回呼ばれた結果そうなっています")
	}
	if len(opt.Fields) == 0 || info.Fields == nil {
		// 全部更新.
		info.Fields = opt.Fields
		info.fieldMap = map[string]bool{}
		for _, field := range opt.Fields {
			info.fieldMap[field] = true
		}
	} else if 0 < len(info.Fields) {
		// 更新するフィールドを追加.
		for _, field := range opt.Fields {
			if _, exists := info.fieldMap[field]; exists {
				continue
			}
			info.Fields = append(info.Fields, field)
			info.fieldMap[field] = true
		}
	}
	if opt.SavedTask != nil {
		info.SavedTasks = append(info.SavedTasks, opt.SavedTask)
	}
	return nil
}

/**
 * 削除するモデルに追加.
 */
func (self *ModelManager) SetDelete(model interface{}) error {
	// 識別子.
	key := self.makeKey(model)
	if d, _ := self.reservedDelModels[key]; d != nil {
		// 設定済み.
		return nil
	}
	// 保存設定されていないかを確認.
	info, _ := self.reservedSaveInfos[key]
	if info != nil {
		return lib_error.NewAppError(lib_error.DefaultErrorCode, "保存予約に入ってる:%s", key)
	}
	// 削除設定.
	self.reservedDelModels[key] = model
	self.reservedDelModelKeys = append(self.reservedDelModelKeys, key)
	return nil
}

/**
 * 書き込み後に実行するタスクを追加.
 */
func (self *ModelManager) AddWriteEndTask(f func(...interface{}) error, args ...interface{}) {
	self.writeEndTasks = append(self.writeEndTasks, &Task{
		F:    f,
		Args: args,
	})
}

/**
 * 予約された書き込みを処理していく.
 */
func (self *ModelManager) WriteAll() error {
	var err error = nil
	if !self.writable {
		return lib_error.WrapError(fmt.Errorf("書き込み不可状態です"))
	}
	// すべてのフィールドを保存したモデル.
	savedModels := []interface{}{}
	// フィールド指定して更新したモデル.
	updatedModels := []interface{}{}
	// 削除したモデル.
	deletedModels := []interface{}{}
	// 保存.
	for _, key := range self.reservedSaveModelKeys {
		info := self.reservedSaveInfos[key]
		if info == nil {
			continue
		}
		modelInfo := self.getInfo(info.Model)
		if info.ForceInsert {
			// insert.
			err = self.db.Create(info.Model).Error
			savedModels = append(savedModels, info.Model)
		} else if 0 < len(info.Fields) {
			// フィールドを指定してupdate.
			db := modelInfo.QueryByModel(self.db, info.Model)
			modelValue := reflect.ValueOf(info.Model)
			if modelValue.Kind() == reflect.Ptr {
				modelValue = reflect.Indirect(modelValue)
			}
			updates := make(map[string]interface{}, len(info.Fields))
			for _, fieldName := range info.Fields {
				field := modelInfo.StructFields[fieldName]
				updates[field.DBName] = modelValue.FieldByName(fieldName).Interface()
			}
			err = db.Updates(updates).Error
			updatedModels = append(updatedModels, info.Model)
		} else if modelInfo.HasCustomTypes() {
			// DBに存在しているかを確認.
			db := modelInfo.QueryByModel(self.db, info.Model)
			cnt := 0
			err = db.Count(&cnt).Error
			if err != nil {
				return lib_error.WrapError(err)
			} else if cnt == 0 {
				// insert.
				err = self.db.Create(info.Model).Error
			} else {
				// update.
				pfMap := make(map[string]bool, len(modelInfo.PrimaryFields))
				for _, pf := range modelInfo.PrimaryFields {
					pfMap[pf.DBName] = true
				}
				modelValue := reflect.Indirect(reflect.ValueOf(info.Model))
				updates := make(map[string]interface{}, len(info.Fields))
				for _, field := range modelInfo.StructFields {
					if isPf, _ := pfMap[field.DBName]; isPf {
						continue
					}
					updates[field.DBName] = modelValue.FieldByName(field.Name).Interface()
				}
				err = db.Updates(updates).Error
			}
			savedModels = append(savedModels, info.Model)
		} else {
			// すべてのフィールドを更新.
			err = self.db.Save(info.Model).Error
			savedModels = append(savedModels, info.Model)
		}
		if err != nil {
			return lib_error.WrapError(err)
		}
		// 書き込み後のタスク.
		for _, task := range info.SavedTasks {
			err = task.F(task.Args...)
			if err != nil {
				return lib_error.WrapError(err)
			}
		}
	}
	// 削除.
	for _, key := range self.reservedDelModelKeys {
		model, _ := self.reservedDelModels[key]
		if model == nil {
			continue
		}
		err = self.db.Delete(model).Error
		if err != nil {
			return lib_error.WrapError(err)
		}
		deletedModels = append(deletedModels, model)
	}
	self.reservedSaveModelKeys = []string{}
	self.reservedSaveInfos = map[string]*ReservedInfo{}
	self.reservedDelModelKeys = []string{}
	self.reservedDelModels = map[string]interface{}{}

	// ModelManagerがガメているモデルを破棄.
	self.modelInfos = map[string]*ModelInfo{}
	// 書き込み後のタスク.
	for _, task := range self.writeEndTasks {
		err = task.F(task.Args...)
		if err != nil {
			return lib_error.WrapError(err)
		}
	}
	// キャッシュを操作.
	err = self.SaveModelsToCache(savedModels, self.CacheTTL, true)
	if err != nil {
		return lib_error.WrapError(err)
	}
	err = self.DeleteModelsFromCache(updatedModels)
	if err != nil {
		return lib_error.WrapError(err)
	}
	err = self.DeleteModelsFromCache(deletedModels)
	if err != nil {
		return lib_error.WrapError(err)
	}

	return nil
}

/**
 * キャッシュにモデルを保存.
 */
func (self *ModelManager) SaveModelsToCache(models []interface{}, ttl int64, update bool) error {
	var err error = nil
	masterTables := map[string]bool{}
	for _, model := range models {
		info := self.getInfo(model)
		if info.GetIsMaster(model) {
			// マスターデータはlocalcache.
			masterTables[info.TableName] = true
			continue
		} else if self.redis == nil {
			return nil
		}
		// トランデータはredisへ.
		key := self.makeKey(model)
		if update {
			// 更新の場合は上書き.
			err = self.redis.Set(key, model)
		} else {
			// 更新ではない場合は無かったときだけ.
			err = self.redis.SetNx(key, model)
		}
		if err == nil {
			err = self.redis.Expire(key, ttl)
		}
		if err != nil {
			return lib_error.WrapError(err)
		}
		self.log("SaveModelsToCache:%s", key)
	}
	if 0 < len(masterTables) {
		// マスターデータのキーを削除.
		if self.redis != nil {
			cacheKeys := make([]interface{}, len(masterTables))
			index := 0
			for tableName, _ := range masterTables {
				cacheKeys[index] = fmt.Sprintf("mastermodel_keyset::%s", tableName)
				index++
			}
			err = self.redis.Del(cacheKeys...)
			if err != nil {
				return lib_error.WrapError(err)
			}
		}
		// ローカルキャッシュを削除.
		err = lib_redis.DeleteLocalCache(self.redis)
		if err != nil {
			return lib_error.WrapError(err)
		}
	}
	return nil
}

/**
 * キャッシュからモデルを削除.
 */
func (self *ModelManager) DeleteModelsFromCache(models []interface{}) error {
	var err error = nil
	masterTables := map[string]bool{}
	keys := []interface{}{}
	for _, model := range models {
		info := self.getInfo(model)
		if info.GetIsMaster(model) {
			// マスターデータはlocalcacheを削除.
			masterTables[info.TableName] = true
			continue
		} else if self.redis == nil {
			return nil
		}
		// トランデータはredisにある.
		keys = append(keys, self.makeKey(model))
	}
	if 0 < len(keys) {
		err = self.redis.Del(keys...)
		if err != nil {
			return lib_error.WrapError(err)
		}
		self.log("DeleteModelsFromCache:%v", keys)
	}
	if 0 < len(masterTables) {
		// マスターデータのキーを削除.
		if self.redis != nil {
			cacheKeys := make([]interface{}, len(masterTables))
			index := 0
			for tableName, _ := range masterTables {
				cacheKeys[index] = fmt.Sprintf("mastermodel_keyset::%s", tableName)
				index++
			}
			err = self.redis.Del(cacheKeys...)
			if err != nil {
				return lib_error.WrapError(err)
			}
		}
		// ローカルキャッシュを削除.
		err = lib_redis.DeleteLocalCache(self.redis)
		if err != nil {
			return lib_error.WrapError(err)
		}
	}
	return nil
}
