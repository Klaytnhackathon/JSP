package db

import (
	"supports/models"
	"time"
	"log"
	"os"
)

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

func GetItems(queryString *map[string]string) ([]models.Support, error) {
	db, err := GetConnection()
	defer db.Close()

	if err != nil {
		errorLogger.Println(err.Error())
		return nil, err
	}

	db = db.Order("created_at desc")

	if (*queryString)["petition_id"] != "" {
		db = db.Where("petition_id = ?", (*queryString)["petition_id"])
	}

	if (*queryString)["signer_id"] != "" {
		db = db.Where("signer_id = ?", (*queryString)["signer_id"])
	}

	supports := []models.Support{}
	if (*queryString)["limit"] == "" {
		db.Find(&supports)
	} else {
		db.Limit((*queryString)["limit"]).Find(&supports)
	}

	return supports, nil
}

func GetTotalCount() (uint, error) {
	db, err := GetConnection()
	defer db.Close()

	if err != nil {
		errorLogger.Println(err.Error())
		return 0, err
	}

	count := (uint)(0)
	db.Table("supports").Count(&count)

	return count, nil
}

func PutItem(support *models.Support) (*models.Support, error) {
	db, err := GetConnection()
	defer db.Close()

	if err != nil {
		errorLogger.Println(err.Error())
		return nil, err
	}

	db.AutoMigrate(&models.Support{})

	support.CreatedAt = time.Now()
	support.UpdatedAt = time.Now()

	temp := models.Support{}
	db.FirstOrCreate(&temp, support)

	return support, nil
}
