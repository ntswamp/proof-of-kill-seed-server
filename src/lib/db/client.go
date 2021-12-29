package lib_db

import (
	lib_redis "app/src/lib/redis"

	"github.com/jinzhu/gorm"
)

type Client struct {
	db           *gorm.DB
	transaction  *gorm.DB
	redis        *lib_redis.Client
	modelManager *ModelManager
}

// インスタンス作成.
func NewClient(db *gorm.DB, redis *lib_redis.Client) *Client {
	return &Client{
		db:           db,
		redis:        redis,
		transaction:  nil,
		modelManager: nil,
	}
}

// 閉じる.
func (self *Client) Close() {
	self.RollbackTransaction()
	self.db.Close()
}

// DBのコネクション情報を取得.
func (self *Client) GetDB() *gorm.DB {
	if self.transaction != nil {
		return self.transaction
	}
	return self.db
}

// キャッシュ用redisクライアントを取得.
func (self *Client) GetRedis() *lib_redis.Client {
	return self.redis
}

// トランザクション内かどうか.
func (self *Client) GetIsInTransaction() bool {
	return self.transaction != nil
}

// モデルマネージャを取得.
func (self *Client) GetModelManager() *ModelManager {
	if self.modelManager == nil {
		self.modelManager = NewModelManager(self.GetDB(), self.redis, self.GetIsInTransaction())
	}
	return self.modelManager
}

// トランザクション処理開始.
func (self *Client) StartTransaction() *gorm.DB {
	if self.transaction != nil {
		// 重複.ここに来るのは設計ミス.クリティカルすぎるのでpanic.
		self.RollbackTransaction()
		panic("Transactions are duplicated.")
	}
	self.transaction = self.db.Begin()
	self.modelManager = nil
	return self.transaction
}

// トランザクション処理終了.
func (self *Client) CommitTransaction() error {
	if self.transaction == nil {
		return nil
	}
	err := self.transaction.Commit().Error
	if err != nil {
		self.RollbackTransaction()
	}
	self.transaction = nil
	self.modelManager = nil
	return err
}

// トランザクション処理をロールバックして終了.
func (self *Client) RollbackTransaction() error {
	if self.transaction == nil {
		return nil
	}
	err := self.transaction.Rollback().Error
	self.transaction = nil
	self.modelManager = nil
	return err
}
