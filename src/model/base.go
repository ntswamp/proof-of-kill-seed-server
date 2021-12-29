package model

import (
	lib_db "app/src/lib/db"
	"time"
)

type ModelInterface interface {
	GetIsMaster() bool
}
type BaseModel struct {
}

func (self BaseModel) GetIsMaster() bool {
	return false
}

type BaseMaster struct {
}

func (self BaseMaster) GetIsMaster() bool {
	return true
}

type Idol struct {
	BaseModel
	Id          uint32 `gorm:"type:serial;primary_key"`
	Name        string
	Age         uint `gorm:"type:smallint;not null"`
	Personality uint `gorm:"type:smallint;not null"`
}

type ChatLog struct {
	BaseModel
	IdolId       uint32 `gorm:"type:integer;primary_key;index"`
	UserId       uint32 `gorm:"type:integer"`
	UserInput    string
	IdolResponse string
	CreatedAt    time.Time `gorm:"type:timestamp"`
}

type PuzzleSaleTransaction struct {
	BuyerEthAccount string
	PuzzleTokenId   string
	TxHash          string
	CreatedAt       time.Time `gorm:"type:timestamp"`
}

type BlockchainWatcher struct {
	BaseModel
	CurrencyType           int   `gorm:"type:smallint;not null;primary_key"`
	LastCheckedBlockNumber int64 `gorm:"type:bigint;not null"`
}

type Puzzle struct {
	TokenId       string
	CreatorUserId uint64
	OwnerUserId   uint64 `gorm:"type:integar;index"`
	Board         string
	Type          uint8     `gorm:"type:smallint"` //1st party, user-made
	Vote          int32     `gorm:"type:integar"`
	CreatedAt     time.Time `gorm:"type:timestamp"`
}

func (self *Puzzle) SetOwner(db *lib_db.Client, ownerUserId uint64) {

}

func (self *Puzzle) Save(db *lib_db.Client) error {
	return nil
}

type RememberMeToken struct {
	BaseModel
	Hash      string    `gorm:"primary_key"`
	UserId    uint64    `gorm:"index"`
	ExpiresAt time.Time `gorm:"type:timestamp"`
}
