package routers

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	smartVoiceTest "task-go/apis/services/smartvoice"
	"task-go/models"
	"task-go/pkg/app"
	"task-go/pkg/e"
	pkgFile "task-go/pkg/file"
	"task-go/pkg/logf"
	"task-go/pkg/util/const"
	groupViews "task-go/routers/views/group"
	projectViews "task-go/routers/views/project"
	serverViews "task-go/routers/views/server"
	skillViews "task-go/routers/views/skill"
	taskViews "task-go/routers/views/task"
	typeViews "task-go/routers/views/type"
	"time"
)

const (
	skill = "skill"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), initLog, cors()) //logging handler for catch all exception

	r.GET("/download", func(c *gin.Context) {
		fileName := c.Query("filename")
		if fileName == "" {
			c.Redirect(http.StatusFound, "/404")
			return
		}
		//打开文件
		f, errByOpenFile := os.Open(fileName)
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {

			}
		}(f)
		//非空处理
		if errByOpenFile != nil {
			c.Redirect(http.StatusFound, "/404")
			return
		}
		list := strings.Split(fileName, "/")
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", "attachment; filename="+list[len(list)-1])
		c.Header("Content-Transfer-Encoding", "binary")
		c.File(fileName)
		return
	})

	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			app.ErrorResp(c, 500, err.Error())
			return
		}
		filename := time.Now().Format("20060102-15-04-05") + file.Filename
		// 上传文件至指定目录
		var once sync.Once
		once.Do(func() {
			dir, err := os.Getwd()
			if err != nil {
				logf.Error(err)
			}
			src := dir + "/upload/"
			err = pkgFile.IsNotExistMkDir(src)
			if err != nil {
				return
			}
		})
		if err := c.SaveUploadedFile(file, "./upload/"+filename); err != nil {
			fmt.Println(err)
		}
		app.SuccessResp(c, filename)
	})

	r.POST("/export", models.ExportExcel)

	//v1
	apiV1 := r.Group("/api/v1")
	apiProjectV1 := apiV1.Group("/project")
	apiProjectV1.Use()
	{
		apiProjectV1.DELETE("/:id", projectViews.DeleteProject)
		apiProjectV1.GET("/:id", projectViews.GetProject)
		apiProjectV1.POST("", projectViews.AddProject)
		apiProjectV1.PUT("/:id", projectViews.UpdateProject)
		apiProjectV1.GET("", projectViews.ListProject)
	}
	apiGroupV1 := apiV1.Group("/group")
	apiGroupV1.Use()
	{
		apiGroupV1.DELETE("/:id", groupViews.DeleteGroup)
		apiGroupV1.GET("/:id", groupViews.GetGroup)
		apiGroupV1.POST("", groupViews.AddGroup)
		apiGroupV1.PUT("/:id", groupViews.UpdateGroup)
		apiGroupV1.GET("", groupViews.ListGroup)
	}
	apiServerV1 := apiV1.Group("/server")
	apiServerV1.Use()
	{
		apiServerV1.DELETE("/:id", serverViews.DeleteServer)
		apiServerV1.GET("/:id", serverViews.GetServer)
		apiServerV1.POST("", serverViews.AddServer)
		apiServerV1.PUT("/:id", serverViews.UpdateServer)
		apiServerV1.GET("", serverViews.ListServer)
	}
	apiTypeV1 := apiV1.Group("/type")
	apiTypeV1.Use()
	{
		apiTypeV1.DELETE("/:id", typeViews.DeleteType)
		apiTypeV1.GET("/:id", typeViews.GetType)
		apiTypeV1.POST("", typeViews.AddType)
		apiTypeV1.PUT("/:id", typeViews.UpdateType)
		apiTypeV1.GET("", typeViews.ListType)
	}
	apiPlanV1 := apiV1.Group("/plan")
	apiPlanV1.Use()
	{
		apiPlanV1.DELETE("/:id", taskViews.DeletePlan)
		apiPlanV1.GET("/:id", taskViews.GetPlan)
		apiPlanV1.POST("", taskViews.AddPlan)
		apiPlanV1.PUT("/:id", taskViews.UpdatePlan)
		apiPlanV1.GET("", taskViews.ListPlan)

		apiPlanV1.POST("/run", taskViews.RunPlanOrEndPlan)
		apiPlanV1.GET("/run", taskViews.ListInstance)
	}
	apiCasesV1 := apiV1.Group("/cases")
	apiCasesV1.Use()
	{
		apiCasesV1.GET("/:type", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				skillViews.ListSkill(context)
			}
		})
		apiCasesV1.DELETE("/:type/:id", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				skillViews.RemoveSkill(context)
			}
		})
		apiCasesV1.GET("/:type/:id", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				skillViews.DetailSkill(context)
			}
		})
		apiCasesV1.POST("/:type", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				skillViews.AddSkill(context)
			}
		})
		apiCasesV1.PUT("/:type/:id", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				skillViews.UpdateSkill(context)
			}
		})
		apiCasesV1.POST("/:type/import/excel", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				skillViews.ImportSkill(context)
			}
		})
		apiCasesV1.GET("/:type/count/:column", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				skillViews.GetSkillCaseCountByColumn(context)
			}
		})
		apiCasesV1.GET("/:type/total/weekly", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				skillViews.GetSkillCaseCountByWeek(context)
			}
		})
	}

	apiReportV1 := apiV1.Group("/reports")
	apiReportV1.Use()
	{
		apiReportV1.POST("/find", models.MongoListFuncFind)
		apiReportV1.POST("/aggregate", models.MongoListFuncAggregate)
		apiReportV1.POST("/find/export", models.MongoListAndExportFunc)
		apiReportV1.PUT("/update", models.MongoUpdateFunc)

		apiReportV1.POST("", models.MongoListFuncFind)
		apiReportV1.POST("/export", models.MongoListAndExportFunc)
		apiReportV1.PUT("", models.MongoUpdateFunc)
	}

	apiTestV1 := apiV1.Group("/test")
	apiTestV1.Use()
	{
		apiTestV1.POST("/smartvoice/:addr", smartVoiceTest.SmartVoiceTestOne)
	}
	return r
}

type Interface interface {
	DeepCopy() interface{}
}

func Copy(src interface{}) interface{} {
	if src == nil {
		return nil
	}
	original := reflect.ValueOf(src)
	cpy := reflect.New(original.Type()).Elem()
	copyRecursive(original, cpy)

	return cpy.Interface()
}

func copyRecursive(src, dst reflect.Value) {
	if src.CanInterface() {
		if copier, ok := src.Interface().(Interface); ok {
			dst.Set(reflect.ValueOf(copier.DeepCopy()))
			return
		}
	}

	switch src.Kind() {
	case reflect.Ptr:
		originalValue := src.Elem()

		if !originalValue.IsValid() {
			return
		}
		dst.Set(reflect.New(originalValue.Type()))
		copyRecursive(originalValue, dst.Elem())

	case reflect.Interface:
		if src.IsNil() {
			return
		}
		originalValue := src.Elem()
		copyValue := reflect.New(originalValue.Type()).Elem()
		copyRecursive(originalValue, copyValue)
		dst.Set(copyValue)

	case reflect.Struct:
		t, ok := src.Interface().(time.Time)
		if ok {
			dst.Set(reflect.ValueOf(t))
			return
		}
		for i := 0; i < src.NumField(); i++ {
			if src.Type().Field(i).PkgPath != "" {
				continue
			}
			copyRecursive(src.Field(i), dst.Field(i))
		}

	case reflect.Slice:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			copyRecursive(src.Index(i), dst.Index(i))
		}

	case reflect.Map:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeMap(src.Type()))
		for _, key := range src.MapKeys() {
			originalValue := src.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			copyRecursive(originalValue, copyValue)
			copyKey := Copy(key.Interface())
			dst.SetMapIndex(reflect.ValueOf(copyKey), copyValue)
		}

	default:
		dst.Set(src)
	}
}

func initLog(c *gin.Context) {
	startTime := time.Now()      // 开始时间
	path := c.Request.RequestURI // 请求路由

	// 排除文件上传的请求体打印
	isFormData := strings.Contains(c.Request.Header.Get("Content-Type"), "multipart/form-data")
	// requestBody
	var requestBody []byte
	if !isFormData {
		requestBody, _ = c.GetRawData()
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		c.Set("requestBody", string(requestBody))
	}

	//处理请求
	c.Next()
	// 处理结果
	result, exists := c.Get(utilconst.LogResponse)
	if exists {
		result = result.(*app.Response)
	}

	// 执行时间
	latencyTime := time.Since(startTime)
	// 请求方式
	reqMethod := c.Request.Method
	// http状态码
	statusCode := c.Writer.Status()
	// 请求IP
	clientIP := c.ClientIP()
	//token := c.GetHeader(tool.HeaderToken)
	// 日志格式
	logf.InfoWithFields(logrus.Fields{
		"req_body":     string(requestBody),
		"http_code":    statusCode,
		"latency_time": fmt.Sprintf("%13v", latencyTime),
		"ip":           clientIP,
		"method":       reqMethod,
		"path":         path,
		"result":       result,
		"msg":          reqMethod,
	})
}

//跨域中间件
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, "+utilconst.HeaderToken)
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type, "+utilconst.HeaderToken)
			c.Header("Access-Control-Allow-Credentials", "false")
			c.Set("content-type", "application/json")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func Validation(req interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		//深拷贝
		copyReq := Copy(req)
		err := app.BindAndValid(c, copyReq)
		if err != nil {
			app.ErrorResp(c, e.InvalidParams, err.Error())
			logf.Debug("Validation", err.Error())
			c.Abort()
			return
		}
		c.Set(utilconst.REQUEST_KEY, copyReq)
	}
}
