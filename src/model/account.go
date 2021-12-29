package model

import (
	lib_db "app/src/lib/db"
	"fmt"
	"time"
)

type Account struct {
	BaseModel
	Wallet       string `gorm:"index;not null;primary_key"`
	Name         string
	language     int       `gorm:"type:smallint"`
	Type         int       `gorm:"type:smallint;not null;default:0"` //0:user 1:admin
	RegisteredAt time.Time `gorm:"type:timestamp"`
}

func ChangeNameOnWallet(db *lib_db.Client, name string, wallet string) error {
	mgr := db.GetModelManager()
	record := &Account{}
	exist, err := mgr.GetModel(record, wallet)
	if err != nil {
		return err
	}
	if exist == nil {
		return fmt.Errorf("wallet not found: %v", wallet)
	}

	//change name
	opt := &lib_db.SaveOptions{
		Fields: []string{"Name"},
	}
	record.Name = name

	//write to db
	db.StartTransaction()
	defer db.RollbackTransaction()

	err = mgr.CachedSave(record, opt)
	if err != nil {
		return err
	}
	err = mgr.WriteAll()
	if err != nil {
		return err
	}
	err = db.CommitTransaction()
	if err != nil {
		return err
	}
	return nil

}

func GetAcountNameByWallet(db *lib_db.Client, wallet string) (string, error) {
	mgr := db.GetModelManager()
	record := &Account{}
	exist, err := mgr.GetModel(record, wallet)
	if err != nil {
		return "", err
	}
	if exist == nil {
		return "", fmt.Errorf("wallet not found: %v", wallet)
	}

	return record.Name, nil
}
