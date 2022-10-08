package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"task-go/pkg/logf"
	"task-go/pkg/mongo"
	"task-go/pkg/setting"
	"time"
)

var ReporterDB *mongo.MongoInfo

func MongoSetup() {
	ReporterDB = setting.ReporterSetting
	ReporterDB.MongoPoolConnect(50)
}

var db, autoMySQLDB *gorm.DB

// MysqlSetup initializes the database instance
func MysqlSetup() {
	var err error
	db, err = gorm.Open(setting.DatabaseSetting.Type, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		setting.DatabaseSetting.User,
		setting.DatabaseSetting.Password,
		setting.DatabaseSetting.Host,
		setting.DatabaseSetting.Name))
	if err != nil {
		logf.Fatal("time.Setup err: %v", err)
	}

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return setting.DatabaseSetting.TablePrefix + defaultTableName
	}

	db.LogMode(setting.DatabaseSetting.Debug)
	db.SingularTable(true)
	// gorm logger
	db.SetLogger(logf.NewGormLogger())
	db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	db.Callback().Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)
	db.Callback().Delete().Replace("gorm:delete", deleteCallback)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	autoMySQLDBSetUp()
}

func autoMySQLDBSetUp() {
	var err error
	autoMySQLDB, err = gorm.Open(setting.DatabaseSetting.Type, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		setting.DatabaseSetting.User,
		setting.DatabaseSetting.Password,
		setting.DatabaseSetting.Host,
		setting.DatabaseSetting.Name))
	if err != nil {
		logf.Fatal("time.Setup err: %v", err)
	}

	gorm.DefaultTableNameHandler = func(autoMySQLDB *gorm.DB, defaultTableName string) string {
		return setting.DatabaseSetting.TablePrefix + defaultTableName
	}

	autoMySQLDB.LogMode(setting.DatabaseSetting.Debug)
	autoMySQLDB.SingularTable(true)
	// gorm logger
	autoMySQLDB.SetLogger(logf.NewGormLogger())
	autoMySQLDB.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	autoMySQLDB.Callback().Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)
	autoMySQLDB.Callback().Delete().Replace("gorm:delete", deleteCallback)
	autoMySQLDB.DB().SetMaxIdleConns(10)
	autoMySQLDB.DB().SetMaxOpenConns(100)
}

// updateTimeStampForCreateCallback will set `CreatedOn`, `ModifiedOn` when creating
func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := time.Now()

		if createTimeField, ok := scope.FieldByName("CreateTime"); ok {
			if createTimeField.IsBlank {
				err := createTimeField.Set(nowTime)
				if err != nil {
					return
				}
			}
		}

		if modifyTimeField, ok := scope.FieldByName("UpdateTime"); ok {
			if modifyTimeField.IsBlank {
				err := modifyTimeField.Set(nowTime)
				if err != nil {
					return
				}
			}
		}
	}
}

// updateTimeStampForUpdateCallback will set `ModifiedOn` when updating
func updateTimeStampForUpdateCallback(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:update_time"); !ok {
		err := scope.SetColumn("UpdateTime", time.Now())
		if err != nil {
			return
		}
	}
}

// deleteCallback will set `DeletedOn` where deleting
func deleteCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var extraOption string
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}

		deletedOnField, hasDeletedOnField := scope.FieldByName("DeletedOn")

		if !scope.Search.Unscoped && hasDeletedOnField {
			scope.Raw(fmt.Sprintf(
				"UPDATE %v SET %v=%v%v%v",
				scope.QuotedTableName(),
				scope.Quote(deletedOnField.DBName),
				scope.AddToVars(time.Now().Unix()),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		} else {
			scope.Raw(fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		}
	}
}

// addExtraSpaceIfExist adds a separator
func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}
