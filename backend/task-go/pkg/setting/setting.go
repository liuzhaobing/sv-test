package setting

import (
	"github.com/go-ini/ini"
	"log"
	"task-go/pkg/mongo"
	"time"
)

type App struct {
	JwtSecret        string
	PageSize         int
	PageSizeLimit    int
	MD5Salt          string // md5 salt
	TokenExpireTime  int64  // token expire (*time.Second)
	TokenRenewalTime int64  // token renewal (*time.Second)
	ImageMaxSize     int
	ImageAllowExts   []string

	// logging related
	RuntimeRootPath string
	LogSavePath     string
	LogSaveName     string
	LogFileExt      string
	TimeFormat      string
}

var AppSetting = &App{}

type Server struct {
	RunMode      string
	HttpPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

var ServerSetting = &Server{}

type Database struct {
	Type        string
	User        string
	Password    string
	Host        string
	Name        string
	TablePrefix string
	Debug       bool
}

var DatabaseSetting = &Database{}

var cfg *ini.File

var ReporterSetting = &mongo.MongoInfo{}

// Setup initialize the configuration instance
func Setup() {
	var err error
	cfg, err = ini.Load("conf/app.ini")
	if err != nil {
		log.Fatalf("setting.Setup, fail to parse 'conf/app.ini': %v", err)
	}

	mapTo("app", AppSetting)
	mapTo("server", ServerSetting)
	mapTo("database", DatabaseSetting)
	mapTo("reporter", ReporterSetting)

	AppSetting.ImageMaxSize = AppSetting.ImageMaxSize * 1024 * 1024
	ServerSetting.ReadTimeout = ServerSetting.ReadTimeout * time.Second
	ServerSetting.WriteTimeout = ServerSetting.WriteTimeout * time.Second
}

// mapTo map section
func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo %s err: %v", section, err)
	}
}
