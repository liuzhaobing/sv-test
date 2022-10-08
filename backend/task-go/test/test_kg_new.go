package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GetInterfaceToString 用途：interface{} 转 string
func GetInterfaceToString(value interface{}) string {
	// interface 转 string
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

// MongoInfo MongoDB连接信息
type MongoInfo struct {
	Addr   string
	DB     string
	client *mongo.Client
}

// MongoPoolConnect 生成连接池
func (m *MongoInfo) MongoPoolConnect(max uint64) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	url := fmt.Sprintf(`mongodb://%s/?connect=direct`, m.Addr)
	mongoOptions := options.Client().ApplyURI(url)
	mongoOptions.SetMaxPoolSize(max)

	var err error
	m.client, err = mongo.Connect(ctx, mongoOptions)
	if err != nil {
		return nil
	}
	return m.client
}

// MongoPoolDisconnect 关闭连接池
func (m *MongoInfo) MongoPoolDisconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	defer func() {
		if err := m.client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

// MongoInsertMany 用途：db.col.insert_many(documents)
func (m *MongoInfo) MongoInsertMany(col string, documents []interface{}, opts ...*options.InsertManyOptions) []interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	collection := m.client.Database(m.DB).Collection(col)
	result, err := collection.InsertMany(ctx, documents, opts...)
	if err != nil {
		log.Fatal(err)
	}
	return result.InsertedIDs
}

// MongoFind 用途：db.col.find(filter)
func (m *MongoInfo) MongoFind(col string, filter interface{}, opts ...*options.FindOptions) (Results []*bson.D) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	collection := m.client.Database(m.DB).Collection(col)
	cur, err := collection.Find(ctx, filter, opts...)
	if err != nil {
		log.Fatal(err)
	}
	Results = m.mongoCursor(ctx, cur)
	return
}

// MongoAggregate 用途：db.col.aggregate()
func (m *MongoInfo) MongoAggregate(col string, filter interface{}, opts ...*options.AggregateOptions) (Results []*bson.D) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	collection := m.client.Database(m.DB).Collection(col)
	cur, err := collection.Aggregate(ctx, filter, opts...)
	if err != nil {
		log.Fatal(err)
	}
	Results = m.mongoCursor(ctx, cur)
	return
}

// mongoCursor 游标 读取查询到的数据
func (m *MongoInfo) mongoCursor(ctx context.Context, cur *mongo.Cursor) (Results []*bson.D) {
	for cur.Next(ctx) {
		var result bson.D
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		Results = append(Results, &result)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	return
}

// MongoCount 用途：db.col.count(filter)
func (m *MongoInfo) MongoCount(col string, filter interface{}, opts ...*options.CountOptions) (count int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	collection := m.client.Database(m.DB).Collection(col)
	count, err := collection.CountDocuments(ctx, filter, opts...)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// MMUE平台登录用户名密码
type mmueLoginInfo struct {
	Username string `json:"username"`
	Password string `json:"pwd"`
}

// MMUE MMUE平台登录信息
type MMUE struct {
	BaseUrl   string
	Token     string
	LoginInfo *mmueLoginInfo
}

// MMUE平台登录方法
func (m *MMUE) mLogin() {
	newHeaderByte, _ := json.Marshal(m.LoginInfo)
	payload := strings.NewReader(string(newHeaderByte))

	req, _ := http.NewRequest("POST", m.BaseUrl+"/mmue/api/login", payload)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	if res.StatusCode == 200 {
		body, _ := ioutil.ReadAll(res.Body)
		bodyString := string(body)
		response := mLoginResponse{}
		err := json.Unmarshal([]byte(bodyString), &response)
		if err != nil {
			return
		}

		m.Token = response.Data.Token
	}

	defer res.Body.Close()
}

// MMUE平台登录接口响应结构体
type mLoginResponse struct {
	Code int `json:"code"`
	Data struct {
		Data struct {
			UserId     int         `json:"user_id"`
			UserName   string      `json:"user_name"`
			UserPower  string      `json:"user_power"`
			TenantId   string      `json:"tenant_id"`
			TenantName string      `json:"tenant_name"`
			TenantLogo string      `json:"tenant_logo"`
			IsRocuser  string      `json:"is_rocuser"`
			LibValue   interface{} `json:"lib_value"`
			AgentId    interface{} `json:"agent_id"`
		} `json:"data"`
		Token string `json:"token"`
	} `json:"data"`
	Msg    string `json:"msg"`
	Status bool   `json:"status"`
}

// HTTPReqInfo http请求信息结构体
type HTTPReqInfo struct {
	Method  string
	Url     string
	Payload io.Reader
}

// MMUE平台发起http请求
func (m *MMUE) mRequest(mReq HTTPReqInfo) []byte {
	if m.Token == "" {
		if 1 == 1 {
			//m.BaseUrl = "https://mmue.region-dev-1.service.iamidata.com"
			m.BaseUrl = "https://mmue-dit87.harix.iamidata.com" // dit 环境
		}
		m.mLogin()
	}
	if 1 == 1 {
		//m.BaseUrl = "http://172.16.13.160:8696"
		//m.BaseUrl = "http://172.16.13.162:8696"
		m.BaseUrl = "http://172.16.23.85:31917" // dit 环境
		mReq.Url = "/kgqa/v1/chat"
	}
	req, _ := http.NewRequest(mReq.Method, m.BaseUrl+mReq.Url, mReq.Payload)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", m.Token)
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("Origin", m.BaseUrl)
	req.Header.Add("Referer", m.BaseUrl+"/app/client")

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		m.mLogin()
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(res.Body)

	body, _ := ioutil.ReadAll(res.Body)
	return body
}

// 知识图谱会话接口
func (m *MMUE) mChat(spaces, query string) *chatResponse {
	//spaces [{"space_name":"common_kg"},{"space_name":"shici_1549937929118056448"}]
	r := HTTPReqInfo{
		Method:  "POST",
		Url:     "/graph/kgqa/v1/chat",
		Payload: strings.NewReader(fmt.Sprintf(`{"spaces": %s, "question": "%s"}`, spaces, query)),
	}
	var c chatResponse
	err := json.Unmarshal(m.mRequest(r), &c)
	if err != nil {
		return nil
	}
	return &c
}

// 知识图谱会话响应结构体
type chatResponse struct {
	Code int `json:"code"`
	Data struct {
		Type       string `json:"@type"`
		EntityName string `json:"entity_name"`
		Disambi    string `json:"disambi"`
		Answer     string `json:"answer"`
		Attr       struct {
			Describ string `json:"describ"`
		} `json:"attr"`
		Source  string `json:"source"`
		TraceId string `json:"trace_id"`
	} `json:"data"`
}

// 算法接口请求体
type kgqaRequest struct {
	Spaces []struct {
		SpaceName string `json:"space_name"`
		Priority  int    `json:"priority"`
	} `json:"spaces"`
	Question string `json:"question"`
	TraceId  string `json:"trace_id"`
}

// 算法接口响应体
type kgqaResponse struct {
	Agentid string `json:"agentid"`
	Traceid string `json:"traceid"`
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Question                 string   `json:"question"`
		EntityName               []string `json:"entity_name"`
		EntityDisambiguationName []string `json:"entity_disambiguation_name"`
		Predicates               []struct {
			PredicateName string `json:"predicate_name"`
			PredicateType string `json:"predicate_type"`
		} `json:"predicates"`
		Properties []string `json:"properties"`
		Answer     string   `json:"answer"`
		ChatSpace  struct {
			SpaceName string `json:"space_name"`
			Priority  int    `json:"priority"`
		} `json:"chat_space"`
	} `json:"result"`
	Version  string `json:"version"`
	Costtime string `json:"costtime"`
}

// 知识图谱会话 功能调试Function
func unitTestForChat() {
	c := &MMUE{
		BaseUrl: "http://10.11.35.104:4000",
		LoginInfo: &mmueLoginInfo{
			Username: "lisha",
			Password: "123456",
		},
	}
	res := c.mChat(fmt.Sprintf(`[{"space_name":"common_kg"},{"space_name":"shici_1549937929118056448"}]`), "周杰伦的母亲是谁")
	jsonByte, _ := json.Marshal(res)
	jsonString := string(jsonByte)
	fmt.Println(jsonString)
}

type BaseConfig struct {
	IsFeiShu   bool
	FeiShuAddr string
	IsCrontab  bool
	CrontabStr string
	IsExcel    bool
}

// MMUETaskConfig 知识图谱会话自动化测试任务 配置文件
type MMUETaskConfig struct {
	*BaseConfig
	*MongoInfo                       //MongoDB连接信息
	*MMUE                            //MMUE平台连接信息
	Reporter              *MongoInfo //用于存放&读取测试结果的MongoDB数据库连接
	ChanNum               int        //期望执行的并发数
	CaseNum               int64      //每一批次执行的case总数
	TaskName              string     //测试任务名称
	IsRandom              bool       //是否随机测试
	CType                 int64      //单跳or两跳
	IsContinue            bool       //断点续传
	ContinuePose          int64      //断点位置
	ContinueJobInstanceId string     //断点任务job_instance_id
	TemplateJson          string
	Spaces                string
}

type MMUETask struct {
	TaskConfig      *MMUETaskConfig       //任务配置文件
	MongoConnection *MongoInfo            //创建的MongoDB对象
	caseListRl      []*bson.D             //所选取的用例实体关系清单
	req             []*MMUETaskReq        //所有请求集合
	RespChan        chan *MMUETaskOneResp //响应信息
	ResultsLog      string                //测试活动日志
	ResultsSummary  string                //测试报告总结
	mu              sync.Mutex
	ChanNum         int    //实际并发数
	Pose            int64  //当前即将加载的用例编号
	TotalCase       int64  //总体可以测试的用例数
	JobInstanceId   string //用于唯一标识本次运行的任务
}

// MMUETaskReq 一次请求所需要的内容
type MMUETaskReq struct {
	Query           string      //组装的query
	ExpectAnswer    string      //期望的answer
	InfoEntityRLIDS interface{} //查询 entity_rl表的结果
}

// MMUETaskRes 一次请求所返回的内容
type MMUETaskRes struct {
	AnswerString string
	AnswerJson   string
	ExecuteTime  int64
	Source       string
	TraceId      string
}

type jsonTime struct {
	time.Duration
}

// MMUETaskOneResp 单次请求的所有信息
type MMUETaskOneResp struct {
	Req     *MMUETaskReq // 单次测试请求信息
	Res     *MMUETaskRes // 单次测试响应信息
	IsPass  bool         // 单次测试测试结果
	EdgCost jsonTime
}

func (m *MMUETask) chanClose() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ChanNum <= 1 {
		close(m.RespChan)
	} else {
		m.ChanNum--
	}
}

// 准备连接MongoDB数据库 连接池
func (m *MMUETask) prePareMongoPool() {
	m.MongoConnection = m.TaskConfig.MongoInfo
	m.MongoConnection.MongoPoolConnect(25)

	m.TaskConfig.Reporter.MongoPoolConnect(25)

	if m.TaskConfig.IsContinue { // 先看是否需要断点续传
		if m.TaskConfig.ContinueJobInstanceId != "" { // 再看下断点任务有没有指定
			m.JobInstanceId = m.TaskConfig.ContinueJobInstanceId
		} else { // 没有指定就自己去数据库找 最新的
			res := m.TaskConfig.Reporter.MongoFind(kgResultsTable, bson.M{}, options.Find().SetSort(bson.M{"execute_time": -1}).SetSkip(0).SetLimit(1))
			m.JobInstanceId = GetInterfaceToString(res[0].Map()["job_instance_id"])
		}

		if m.TaskConfig.ContinuePose != 0 { // 最后看下断点开始位置
			m.Pose = m.TaskConfig.ContinuePose
		} else { // 没有指定就自己去数据库找 统计数值
			m.Pose = m.TaskConfig.Reporter.MongoCount(kgResultsTable, bson.M{"job_instance_id": m.JobInstanceId})
		}
	} else { // 不需要断点续传就开始新的任务了
		m.JobInstanceId = uuid.New().String()
	}
}

// 单跳 组装query Request
func (m *MMUETask) fakeQuerySingleStep(entityRl *bson.D) (Req *MMUETaskReq) {
	var infoEID1, infoEID2, infoOTRLID []*bson.D
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		infoEID1 = m.MongoConnection.MongoFind(entityTable, bson.M{"_id": entityRl.Map()["e_id"], "need_audit": false}, options.Find().SetLimit(1))
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		infoEID2 = m.MongoConnection.MongoFind(entityTable, bson.M{"_id": entityRl.Map()["e_id2"], "need_audit": false}, options.Find().SetLimit(1))
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		infoOTRLID = m.MongoConnection.MongoFind(ontologyRLTable, bson.M{"_id": entityRl.Map()["ot_rl_id"]}, options.Find().SetLimit(1))
		wg.Done()
	}()
	wg.Wait()

	if infoEID1 != nil && infoOTRLID != nil && infoEID2 != nil {
		Req = &MMUETaskReq{
			Query:           GetInterfaceToString(infoEID1[0].Map()["name"]) + "的" + GetInterfaceToString(infoOTRLID[0].Map()["name"]),
			ExpectAnswer:    GetInterfaceToString(infoEID2[0].Map()["name"]),
			InfoEntityRLIDS: entityRl.Map()["_id"],
		}
	}
	return
}

// 单跳 抽取关系 查看关系表中总共有多少条满足测试条件的关系
func (m *MMUETask) prePareNextBatchCount() {
	m.TotalCase = m.MongoConnection.MongoCount(entityRLTable, entityRLFilter)
}

// 单跳 从关系表中按照批次抽取下一批关系 组成用例 并清空上一批次数据的内存占用
func (m *MMUETask) oneStepPrePareNextBatchCases() []*MMUETaskReq {
	Log := fmt.Sprintf("%s 开始准备用例...\n", time.Now().Format("2006-01-02-15-04-05"))
	m.ResultsLog = Log
	fmt.Println(Log)
	if m.Pose <= m.TotalCase {
		// 抽取关系 从关系表中 分页分批次收集抽取n条关系
		m.caseListRl = m.MongoConnection.MongoFind(entityRLTable, entityRLFilter, options.Find().SetLimit(m.TaskConfig.CaseNum).SetSkip(m.Pose))
		m.Pose = m.Pose + m.TaskConfig.CaseNum

		m.req = nil
		for _, i := range m.caseListRl {
			r := m.fakeQuerySingleStep(i)
			if r != nil {
				m.req = append(m.req, r)
			}
		}
	} else {
		m.caseListRl, m.req = nil, nil
		return m.req
	}

	Log = fmt.Sprintf("%s 有效用例%d条...\n", time.Now().Format("2006-01-02-15-04-05"), len(m.req))
	m.ResultsLog = m.ResultsLog + Log
	fmt.Println(Log)

	m.RespChan = make(chan *MMUETaskOneResp, len(m.req))

	if m.TaskConfig.ChanNum > 0 {
		m.ChanNum = m.TaskConfig.ChanNum
	} else {
		m.ChanNum = 1
	}

	return m.req
}

// 单跳 从关系表中随机抽取num条关系 组成用例
func (m *MMUETask) oneStepPrePareRandomCases() []*MMUETaskReq {
	m.req = nil
	m.caseListRl = nil
	Log := fmt.Sprintf("%s 开始准备用例...\n", time.Now().Format("2006-01-02-15-04-05"))
	m.ResultsLog = Log
	fmt.Println(Log)

	// 抽取关系 从关系表中 随机抽取n条关系
	m.caseListRl = m.MongoConnection.MongoAggregate(entityRLTable, []bson.M{
		{"$sample": bson.M{"size": m.TaskConfig.CaseNum}},
		{"$match": entityRLFilter}})
	for _, i := range m.caseListRl {
		r := m.fakeQuerySingleStep(i)
		if r != nil {
			m.req = append(m.req, r)
		}
	}

	Log = fmt.Sprintf("%s 有效用例%d条...\n", time.Now().Format("2006-01-02-15-04-05"), len(m.req))
	m.ResultsLog = m.ResultsLog + Log
	fmt.Println(Log)

	m.RespChan = make(chan *MMUETaskOneResp, len(m.req))

	if m.TaskConfig.ChanNum > 0 {
		m.ChanNum = m.TaskConfig.ChanNum
	} else {
		m.ChanNum = 1
	}

	return m.req
}

// 两跳 对查询到的数据组进行切片处理
func returnNumSlice(n int, x []*bson.D) []*bson.D {
	if len(x) > n {
		rand.Seed(time.Now().UnixNano())
		q := rand.Intn(len(x) - n)
		x = x[q : q+n]
	}
	return x
}

// 两跳 模板JSON文件结构
type template struct {
	Rl1OntologyName string `json:"rl1_ontology_name"`
	Model           []struct {
		Query           string `json:"query"`
		Rl2OntologyName string `json:"rl2_ontology_name"`
	} `json:"model"`
}

// 两跳 从JSON文件中获取所有模板
func (m *MMUETask) readTemplateFromJson(path string) (te []*template) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {

		}
	}(jsonFile)
	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return
	}
	err = json.Unmarshal(jsonData, &te)
	if err != nil {
		return nil
	}
	return te
}

// 两跳 每次读取模板中的一条
func (m *MMUETask) getOneTemplate() (string, string, string) {
	tmpList := m.readTemplateFromJson(m.TaskConfig.TemplateJson)
	if m.TaskConfig.IsRandom {
		rand.Seed(time.Now().UnixNano())
		tmp := tmpList[rand.Intn(len(tmpList))]
		model := tmp.Model[rand.Intn(len(tmp.Model))]
		return tmp.Rl1OntologyName, model.Rl2OntologyName, model.Query
	} else {
		// 不随机 就先返回第一个 后面再看下怎么去处理
		return tmpList[0].Rl1OntologyName, tmpList[0].Model[0].Rl2OntologyName, tmpList[0].Model[0].Query
	}
}

// 两跳 组装query Request
func (m *MMUETask) fakeQueryDoubleStepNew() (Req *MMUETaskReq) {
	// rl1OntologyName = "作者"
	// rl2OntologyName = "代表作"
	// model = "的作者有哪些代表作"
	nu := 20
	rl1OntologyName, rl2OntologyName, model := m.getOneTemplate()

	// 在ontology_rl中找作者属性id
	authorRl := m.MongoConnection.MongoFind(ontologyRLTable, bson.M{"name": rl1OntologyName}) // 这儿找到420条数据

	if m.TaskConfig.IsRandom {
		authorRl = returnNumSlice(nu, authorRl)
	}

	for _, a := range authorRl {
		// 根据作者属性id 找第一个三元组关系
		da := m.MongoConnection.MongoFind(entityRLTable, bson.M{"ot_rl_id": a.Map()["_id"]}) // 这儿找到一堆的三元组关系1

		if m.TaskConfig.IsRandom {
			da = returnNumSlice(nu, da)
		}

		for _, b := range da {
			var wg sync.WaitGroup
			var f, q []*bson.D
			wg.Add(1)
			go func() {
				f = m.MongoConnection.MongoFind(entityRLTable, bson.M{"e_id": b.Map()["e_id2"]})
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				// 根据三元组关系找实体1 组装query
				q = m.MongoConnection.MongoFind(entityTable, bson.M{"_id": b.Map()["e_id"], "need_audit": false}, options.Find().SetLimit(1))
				wg.Done()
			}()
			wg.Wait()

			if m.TaskConfig.IsRandom {
				f = returnNumSlice(nu, f)
			}

			//	根据eid2 找第二个三元组关系
			for _, x := range f {
				var kk, n []*bson.D
				wg.Add(1)
				go func() {
					kk = m.MongoConnection.MongoFind(ontologyRLTable, bson.M{"_id": x.Map()["ot_rl_id"]})
					wg.Done()
				}()
				wg.Add(1)
				go func() {
					n = m.MongoConnection.MongoFind(entityTable, bson.M{"_id": x.Map()["e_id2"], "need_audit": false}, options.Find().SetLimit(1))
					wg.Done()
				}()
				wg.Wait()
				if kk != nil && n != nil && q != nil {
					if kk[0].Map()["name"] == rl2OntologyName {
						Req = &MMUETaskReq{
							Query:        GetInterfaceToString(q[0].Map()["name"]) + model,
							ExpectAnswer: GetInterfaceToString(n[0].Map()["name"]),
						}
						return
					}
				}
			}
		}
	}
	return nil
}

// 废弃两跳 组装query Request
func (m *MMUETask) fakeQueryDoubleStep(entityRl *bson.D) (Req *MMUETaskReq) {
	var infoEID1, infoOTRLID, entityRl2, secondEID2, secondOTRLID []*bson.D
	var wg sync.WaitGroup
	// 查询第一跳相关信息
	wg.Add(1)
	go func() {
		infoEID1 = m.MongoConnection.MongoFind(entityTable, bson.M{"_id": entityRl.Map()["e_id"], "need_audit": false}, options.Find().SetLimit(1))
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		infoOTRLID = m.MongoConnection.MongoFind(ontologyRLTable, bson.M{"_id": entityRl.Map()["ot_rl_id"]}, options.Find().SetLimit(1))
		wg.Done()
	}()

	// 使用第一跳的eid2作为第二跳的eid1去查询关系
	wg.Add(1)
	go func() {
		entityRl2 = m.MongoConnection.MongoFind(entityRLTable, bson.M{"e_id": entityRl.Map()["e_id2"]})
		wg.Done()
	}()
	wg.Wait()

	if entityRl2 != nil {
		rand.Seed(time.Now().UnixNano())
		var index, count int
		for {
			index = rand.Intn(len(entityRl2))
			if entityRl2[index].Map()["e_id"] != entityRl2[index].Map()["e_id2"] {
				break
			} else {
				if count < 5 {
					count++
				} else {
					return nil
				}
			}
		}
		wg.Add(1)
		go func() {
			secondOTRLID = m.MongoConnection.MongoFind(ontologyRLTable, bson.M{"_id": entityRl2[index].Map()["ot_rl_id"]}, options.Find().SetLimit(1))
			wg.Done()
		}()
		wg.Add(1)
		go func() {
			secondEID2 = m.MongoConnection.MongoFind(entityTable, bson.M{"_id": entityRl2[index].Map()["e_id2"], "need_audit": false}, options.Find().SetLimit(1))
			wg.Done()
		}()
		wg.Wait()

		if secondEID2 != nil && infoEID1 != nil {
			Req = &MMUETaskReq{
				Query:           GetInterfaceToString(infoEID1[0].Map()["name"]) + "的" + GetInterfaceToString(infoOTRLID[0].Map()["name"]) + "的" + GetInterfaceToString(secondOTRLID[0].Map()["name"]),
				ExpectAnswer:    GetInterfaceToString(secondEID2[0].Map()["name"]),
				InfoEntityRLIDS: entityRl.Map()["_id"],
			}
			return
		}
		return nil
	} else {
		return nil
	}
}

// 两跳 随机组装num条用例
func (m *MMUETask) twoStepPrePareRandomCases() []*MMUETaskReq {
	// 先清空上次的数据 释放内存
	m.req = nil
	m.caseListRl = nil

	Log := fmt.Sprintf("%s 开始准备用例...\n", time.Now().Format("2006-01-02-15-04-05"))
	m.ResultsLog = Log
	fmt.Println(Log)

	var i int64
	for i = 0; i < m.TaskConfig.CaseNum; i++ {
		r := m.fakeQueryDoubleStepNew()
		if r != nil {
			m.req = append(m.req, r)
		} else {
			i--
		}
	}

	Log = fmt.Sprintf("%s 有效用例%d条...\n", time.Now().Format("2006-01-02-15-04-05"), len(m.req))
	m.ResultsLog = m.ResultsLog + Log
	fmt.Println(Log)

	m.RespChan = make(chan *MMUETaskOneResp, len(m.req))

	if m.TaskConfig.ChanNum > 0 {
		m.ChanNum = m.TaskConfig.ChanNum
	} else {
		m.ChanNum = 1
	}
	return m.req
}

// 用例收集器
func (m *MMUETask) recordCase() {
	//writeDB
	var KGResults []interface{}
	for _, resp := range m.req {
		KGResults = append(KGResults, &KGResult{
			JobInstanceId: m.JobInstanceId,
			Question:      resp.Query,
			Answer:        resp.ExpectAnswer,
			TaskName:      m.TaskConfig.TaskName,
		})
	}
	m.TaskConfig.Reporter.MongoInsertMany(kgResultsTable, KGResults)
	Log := fmt.Sprintf("%s 收集到有效用例%d条...\n", time.Now().Format("2006-01-02-15-04-05"), len(m.req))
	fmt.Println(Log)
}

// 并发执行用例
func (m *MMUETask) run() {
	Log := fmt.Sprintf("%s 开始执行用例...\n", time.Now().Format("2006-01-02-15-04-05"))
	m.ResultsLog = m.ResultsLog + Log
	fmt.Println(Log)

	ReqChan := make(chan *MMUETaskReq)
	for i := 0; i < m.ChanNum; i++ {
		go func(i int, v chan *MMUETaskReq) {
			defer m.chanClose()

			for req := range v {
				res := m.call(req)
				m.RespChan <- res
			}
		}(i, ReqChan)
	}

	for _, req := range m.req {
		ReqChan <- req
	}
	close(ReqChan)

	//writeDB
	var KGResults []interface{}
	for resp := range m.RespChan {
		KGResults = append(KGResults, &KGResult{
			JobInstanceId: m.JobInstanceId,
			Question:      resp.Req.Query,
			Answer:        resp.Req.ExpectAnswer,
			ActAnswer:     resp.Res.AnswerString,
			IsPass:        resp.IsPass,
			RespJson:      resp.Res.AnswerJson,
			EntityRlId:    resp.Req.InfoEntityRLIDS,
			EdgCost:       resp.EdgCost.Milliseconds(),
			ExecuteTime:   resp.Res.ExecuteTime,
			TaskName:      m.TaskConfig.TaskName,
			Source:        resp.Res.Source,
			TraceId:       resp.Res.TraceId,
		})
	}
	m.TaskConfig.Reporter.MongoInsertMany(kgResultsTable, KGResults)

	// do summary
	total := m.TaskConfig.Reporter.MongoCount(kgResultsTable, bson.M{"job_instance_id": m.JobInstanceId})
	fail := m.TaskConfig.Reporter.MongoCount(kgResultsTable, bson.M{"job_instance_id": m.JobInstanceId, "is_pass": false})

	costInfo := m.TaskConfig.Reporter.MongoAggregate(kgResultsTable, []bson.M{
		{"$match": bson.M{"job_instance_id": m.JobInstanceId}},
		{"$group": bson.M{
			"_id":     "$job_instance_id",
			"maxCost": bson.M{"$max": "$edg_cost"},
			"minCost": bson.M{"$min": "$edg_cost"},
			"avgCost": bson.M{"$avg": "$edg_cost"},
		}}})

	m.ResultsSummary = fmt.Sprintf("%s：用例统计:%d, 错误数:%d, 正确率:%.4f, 用例并发数:%d", m.TaskConfig.TaskName, total, fail, 1-float32(fail)/float32(total), m.TaskConfig.ChanNum)
	if costInfo != nil {
		m.ResultsSummary = m.ResultsSummary + fmt.Sprintf("\n最大耗时:%d, 最小耗时:%d, 平均耗时:%.2f",
			costInfo[0].Map()["maxCost"],
			costInfo[0].Map()["minCost"],
			costInfo[0].Map()["avgCost"],
		)
	}

	Log = fmt.Sprintf("%s 用例执行完成...\n", time.Now().Format("2006-01-02-15-04-05"))
	m.ResultsLog = m.ResultsLog + Log
	fmt.Println(Log)

	// do FeiShu report
	if m.TaskConfig.FeiShuAddr != "" && m.TaskConfig.IsFeiShu {
		feiShuText := m.ResultsSummary + fmt.Sprintf("\n结果查询：mongodb://%s/%s\n查询命令：db.%s.find({'job_instance_id': '%s'})", m.TaskConfig.Reporter.Addr, m.TaskConfig.Reporter.DB, kgResultsTable, m.JobInstanceId)
		httpToFeiShu(feiShuText, m.TaskConfig.FeiShuAddr)
	}
}

// 单条用例执行详情
func (m *MMUETask) call(req *MMUETaskReq) *MMUETaskOneResp {
	executeTime, _ := strconv.ParseInt(time.Now().Format("20060102150405"), 10, 64)
	Res := &MMUETaskOneResp{
		Req: req,
		Res: &MMUETaskRes{
			AnswerString: "",
			AnswerJson:   "",
			ExecuteTime:  executeTime,
		},
	}
	// do test
	startReq := time.Now()

	res := m.TaskConfig.MMUE.mChat(m.TaskConfig.Spaces, req.Query)
	Res.EdgCost.Duration = time.Now().Sub(startReq)

	jsonByte, _ := json.Marshal(res)
	Res.Res.AnswerJson = string(jsonByte)

	if res != nil {
		Res.Res.AnswerString = res.Data.Answer
		Res.Res.Source = res.Data.Source
		Res.Res.TraceId = res.Data.TraceId

		// do assertion
		if strings.Contains(Res.Res.AnswerString, Res.Req.ExpectAnswer) {
			Res.IsPass = true
		} else {
			Res.IsPass = false
		}
	}
	return Res
}

// KGResult 存储kg测试结果的MongoDB表结构
type KGResult struct {
	JobInstanceId string      `bson:"job_instance_id"` // 唯一标识本次任务的id
	Question      string      `bson:"question"`
	Answer        string      `bson:"answer"`
	ActAnswer     string      `bson:"act_answer"`
	IsPass        bool        `bson:"is_pass"`
	RespJson      string      `bson:"resp_json"`
	EntityRlId    interface{} `bson:"entity_rl_id"`
	EdgCost       int64       `bson:"edg_cost"`
	ExecuteTime   int64       `bson:"execute_time"`
	TaskName      string      `bson:"task_name"`
	Source        string      `bson:"source"`
	TraceId       string      `bson:"trace_id"`
}

// 发FeiShu的Function
func httpToFeiShu(text, url string) {

	method := "POST"
	newString := strings.ReplaceAll(text, "\n", "\\n")
	newString = strings.ReplaceAll(newString, "\t", "\\t")
	payload := strings.NewReader(fmt.Sprintf("{\n    \"msg_type\": \"text\",\n    \"content\": {\n        \"text\": \"%s\"\n    }\n}", newString))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)
	_, err = ioutil.ReadAll(res.Body)

	if err != nil {
		return
	}
}

// 收集测试用例
func (m *MMUETask) testCaseGetter() {
	switch m.TaskConfig.CType {
	case 1:
		if m.TaskConfig.IsRandom {
			// 知识图谱自动化测试 随机关系
			m.prePareMongoPool()
			for {
				m.oneStepPrePareRandomCases()
				m.recordCase()
			}
		} else {
			m.prePareMongoPool()
			m.prePareNextBatchCount()
			// 知识图谱自动化测试 长时间测试
			for m.Pose <= m.TotalCase {
				m.oneStepPrePareNextBatchCases()
				m.recordCase()
			}
		}
	case 2:
		if m.TaskConfig.IsRandom {

			// 先一直循环着吧
			for {
				m.prePareMongoPool()
				m.twoStepPrePareRandomCases()
				m.recordCase()
			}
		} else {
			// 不然呢？
		}
	}
}

func (m *MMUETask) testCaseRunner() {
	switch m.TaskConfig.CType {
	case 1:
		if m.TaskConfig.IsRandom {
			// 知识图谱自动化测试 随机关系
			m.prePareMongoPool()
			for {
				m.oneStepPrePareRandomCases()
				m.run()
			}
		} else {
			m.prePareMongoPool()
			m.prePareNextBatchCount()
			// 知识图谱自动化测试 长时间测试
			for m.Pose <= m.TotalCase {
				m.oneStepPrePareNextBatchCases()
				m.run()
			}
		}
	case 2:
		if m.TaskConfig.IsRandom {

			// 先一直循环着吧
			for {
				m.prePareMongoPool()
				m.twoStepPrePareRandomCases()
				m.run()
			}
		} else {
			// 不然呢？
		}
	}
}

// 从测试报告中拉取数据作为测试用例
func (m *MMUETask) testCasePrepareFromReport(filter interface{}) []*MMUETaskReq {
	// 先清空上次的数据 释放内存
	m.req = nil
	m.caseListRl = nil

	Log := fmt.Sprintf("%s 开始准备用例...\n", time.Now().Format("2006-01-02-15-04-05"))
	m.ResultsLog = Log
	fmt.Println(Log)

	// 抽取用例
	cases := m.TaskConfig.Reporter.MongoFind(kgResultsTable, filter, options.Find().SetLimit(m.TaskConfig.CaseNum).SetSkip(m.Pose))
	m.Pose = m.Pose + m.TaskConfig.CaseNum
	for _, i := range cases {
		r := &MMUETaskReq{
			Query:        GetInterfaceToString(i.Map()["question"]),
			ExpectAnswer: GetInterfaceToString(i.Map()["answer"]),
		}
		if r != nil {
			m.req = append(m.req, r)
		}
	}

	Log = fmt.Sprintf("%s 有效用例%d条...\n", time.Now().Format("2006-01-02-15-04-05"), len(m.req))
	m.ResultsLog = m.ResultsLog + Log
	fmt.Println(Log)

	m.RespChan = make(chan *MMUETaskOneResp, len(m.req))

	if m.TaskConfig.ChanNum > 0 {
		m.ChanNum = m.TaskConfig.ChanNum
	} else {
		m.ChanNum = 1
	}

	return m.req
}

// 从测试报告中拉取数据作为测试用例
func (m *MMUETask) testCaseRunnerFromReport(filter interface{}) {
	m.prePareMongoPool()
	for {
		m.testCasePrepareFromReport(filter)
		if m.req == nil {
			break
		}
		m.run()
	}
}

func unitTestMMUETask() {
	t := &MMUETask{
		TaskConfig: &MMUETaskConfig{
			TaskName: "common_kg_v4单跳用例构造",
			BaseConfig: &BaseConfig{
				IsFeiShu:   true,
				FeiShuAddr: "https://open.feishu.cn/open-apis/bot/v2/hook/e645f4f3-2c4d-4bdc-a5ac-ccf6865740f2",
				//FeiShuAddr: "",
				IsCrontab:  false,
				CrontabStr: "",
				IsExcel:    false,
			},
			ChanNum:               1,                                      // 对话并发数 内部服务器开并发会很卡 所以这里给关掉了
			CaseNum:               10000,                                  // 每一批次选择的case总数
			IsRandom:              true,                                   // true:随机测试CaseNum条用例 false:进行压力测试
			CType:                 1,                                      // 单跳1 两跳2 三跳3
			IsContinue:            true,                                   // 是否开启断点续传
			ContinueJobInstanceId: "92785d77-6102-4bd3-a207-4f41ad0d8d92", // 断点续传任务的job_instance_id，如果不填，系统会自动找最后执行的那次任务的id
			ContinuePose:          0,                                      // 断点续传位置，如果不填，系统会根据job_instance_id已经执行的数量去指定跳过执行数量
			TemplateJson:          "./test_kg_model12.json",
			Spaces:                fmt.Sprintf(`[{"space_name":"common_kg_v4"}]`), //多图空间
			MongoInfo: &MongoInfo{
				Addr: "172.16.23.85:30966",
				DB:   "common_kg_v4",
			},
			MMUE: &MMUE{
				BaseUrl: "https://mmue.region-dev-1.service.iamidata.com",
				LoginInfo: &mmueLoginInfo{
					Username: "liuzhaobing",
					Password: "123456",
				},
			},
			Reporter: &MongoInfo{
				Addr: "root:123456@172.16.23.33:27927",
				DB:   "autotest",
			},
		},
	}
	t.testCaseRunner()
}

func unitTestCasesGetter() {
	t := &MMUETask{
		TaskConfig: &MMUETaskConfig{
			TaskName: "common_kg_v4单跳用例构造",
			BaseConfig: &BaseConfig{
				IsFeiShu:   true,
				FeiShuAddr: "https://open.feishu.cn/open-apis/bot/v2/hook/e645f4f3-2c4d-4bdc-a5ac-ccf6865740f2",
				//FeiShuAddr: "",
				IsCrontab:  false,
				CrontabStr: "",
				IsExcel:    false,
			},
			ChanNum:               1,                            // 对话并发数 内部服务器开并发会很卡 所以这里给关掉了
			CaseNum:               1000,                         // 每一批次选择的case总数
			IsRandom:              true,                         // true:随机测试CaseNum条用例 false:进行压力测试
			CType:                 2,                            // 单跳1 两跳2 三跳3
			IsContinue:            true,                         // 是否开启断点续传
			ContinueJobInstanceId: "20220915common_kg_v4_case2", // 断点续传任务的job_instance_id，如果不填，系统会自动找最后执行的那次任务的id
			ContinuePose:          0,                            // 断点续传位置，如果不填，系统会根据job_instance_id已经执行的数量去指定跳过执行数量
			TemplateJson:          "./test_kg_model12.json",
			Spaces:                fmt.Sprintf(`[{"space_name":"common_kg_v4"}]`), //多图空间
			MongoInfo: &MongoInfo{
				Addr: "172.16.23.85:30966",
				DB:   "common_kg_v4",
			},
			MMUE: &MMUE{
				BaseUrl: "https://mmue.region-dev-1.service.iamidata.com",
				LoginInfo: &mmueLoginInfo{
					Username: "liuzhaobing",
					Password: "123456",
				},
			},
			Reporter: &MongoInfo{
				Addr: "root:123456@172.16.23.33:27927",
				DB:   "autotest",
			},
		},
	}
	t.testCaseGetter()
}

func unitTestMMUETaskFromReport() {
	t := &MMUETask{
		TaskConfig: &MMUETaskConfig{
			TaskName: "common_kg_v4两跳用例构造",
			BaseConfig: &BaseConfig{
				IsFeiShu:   true,
				FeiShuAddr: "https://open.feishu.cn/open-apis/bot/v2/hook/e645f4f3-2c4d-4bdc-a5ac-ccf6865740f2",
				//FeiShuAddr: "https://open.feishu.cn/open-apis/bot/v2/hook/ad39f3a4-4ff2-47e7-9cd0-dcaf22aeb366",
				IsCrontab:  false,
				CrontabStr: "",
				IsExcel:    false,
			},
			ChanNum:               1,    // 对话并发数 内部服务器开并发会很卡 所以这里给关掉了
			CaseNum:               5000, // 每一批次选择的case总数
			IsRandom:              true, // true:随机测试CaseNum条用例 false:进行压力测试
			CType:                 1,    // 单跳1 两跳2 三跳3
			IsContinue:            true, // 是否开启断点续传
			ContinuePose:          0,    // 断点续传位置，如果不填，系统会根据job_instance_id已经执行的数量去指定跳过执行数量
			TemplateJson:          "task/kg/test_kg_model.json",
			Spaces:                fmt.Sprintf(`[{"space_name":"common_kg_v4"}]`), //多图空间
			ContinueJobInstanceId: "929naac1-4144-47c5-9867-36b86124b7ac",
			MongoInfo: &MongoInfo{
				Addr: "root:123456@172.16.23.33:27927",
				DB:   "autotest",
			},
			MMUE: &MMUE{
				BaseUrl: "https://mmue.region-dev-1.service.iamidata.com",
				LoginInfo: &mmueLoginInfo{
					Username: "liuzhaobing",
					Password: "123456",
				},
			},
			Reporter: &MongoInfo{
				Addr: "root:123456@172.16.23.33:27927",
				DB:   "autotest",
			},
		},
	}
	t.testCaseRunnerFromReport(bson.M{"job_instance_id": "20220923common_kg_v4_case2"}) // 单跳数据源
	//t.testCaseRunnerFromReport(bson.M{"job_instance_id": "9994685c-7731-4644-a1f6-8ba84c735d01"}) // 两跳数据源
	//t.testCaseRunnerFromReport(bson.M{"job_instance_id": "545d2895-b632-4b65-bbb5-1c1433326bb2"}) // 单跳并发数据源 new
	//t.testCaseRunnerFromReport(bson.M{"job_instance_id": bson.M{"$in": bson.A{"7054685c-7731-4644-a1f6-8ba84c735d01", "92a2a778-44d5-46c3-97ec-4fc00c2c77e6"}}}) // 两跳并发数据源 old
	//t.testCaseRunnerFromReport(bson.M{"job_instance_id": "b80fa570-0c13-4e04-b920-d1b606c4bd15", "is_pass": true}) // 两跳并发数据源 new
}

func unitTestTwo() {
	t := &MMUETask{
		TaskConfig: &MMUETaskConfig{
			TaskName: "test_kg",
			BaseConfig: &BaseConfig{
				IsFeiShu:   true,
				FeiShuAddr: "https://open.feishu.cn/open-apis/bot/v2/hook/e645f4f3-2c4d-4bdc-a5ac-ccf6865740f2",
				IsCrontab:  false,
				CrontabStr: "",
				IsExcel:    false,
			},
			ChanNum:               1,    // 对话并发数 内部服务器开并发会很卡 所以这里给关掉了
			CaseNum:               2000, // 每一批次选择的case总数
			IsRandom:              true, // true:随机测试CaseNum条用例 false:进行压力测试
			CType:                 1,    // 单跳1 两跳2 三跳3
			IsContinue:            true, // 是否开启断点续传
			ContinueJobInstanceId: "",   // 断点续传任务的job_instance_id，如果不填，系统会自动找最后执行的那次任务的id
			ContinuePose:          0,    // 断点续传位置，如果不填，系统会根据job_instance_id已经执行的数量去指定跳过执行数量
			TemplateJson:          "task/kg/test_kg_model.json",
			Spaces:                fmt.Sprintf(`[{"space_name":"common_kg"},{"space_name":"shici_1549937929118056448"}]`), //多图空间
			MongoInfo: &MongoInfo{
				Addr: "172.16.23.85:30966",
				DB:   "common_kg",
			},
			MMUE: &MMUE{
				BaseUrl: "https://mmue.region-dev-1.service.iamidata.com",
				LoginInfo: &mmueLoginInfo{
					Username: "three",
					Password: "123456",
				},
			},
			Reporter: &MongoInfo{
				Addr: "root:123456@10.12.32.30:27017",
				DB:   "autotest",
			},
		},
	}
	t.prePareMongoPool()
	f := t.fakeQueryDoubleStepNew()
	fmt.Println(f)

}

func main() {
	//unitTestMMUETask()
	//unitTestCasesGetter()
	//unitTestTwo()
	unitTestMMUETaskFromReport()
}

var (
	entityRLFilter  = bson.M{"status": 0}
	entityRLTable   = "entity_rl"
	entityTable     = "entity"
	ontologyRLTable = "ontology_rl"
	kgResultsTable  = "kg_results"
)
