package task

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"task-go/pkg/mongo"
	"time"
)

type KGTaskConfig struct {
	PlatFrom              *MMUE            `json:"plat_from"`
	DataSource            *mongo.MongoInfo `json:"data_source"` // test case data source
	DataReport            *mongo.MongoInfo `json:"data_report"` // test data report
	ChanNum               int64            `json:"chan_num"`
	CaseNum               int64            //每一批次执行的case总数
	TaskName              string           //测试任务名称
	IsRandom              bool             //是否随机测试
	CType                 int64            //单跳or两跳
	IsContinue            bool             //断点续传
	ContinuePose          int64            //断点位置
	ContinueJobInstanceId string           //断点任务job_instance_id
	TemplateJson          string
	Spaces                string
}

type KGTask struct {
	baseTask
	TaskConfig    *KGTaskConfig        //任务配置
	req           []*KGTaskReq         //所有请求集合
	RespChan      chan *KGTaskOnceResp //响应信息
	startTime     time.Time            //请求时间
	endTime       time.Time            //完成时间
	cost          time.Duration        //总计耗时
	mu            sync.Mutex
	ChanNum       int64  //实际并发数
	JobInstanceId string //用于唯一标识本次运行的任务
}

// KGTaskReq 一次请求所需要的内容
type KGTaskReq struct {
	Query           string      //组装的query
	ExpectAnswer    string      //期望的answer
	InfoEntityRLIDS interface{} //查询 entity_rl表的结果
}

// KGTaskRes 一次请求所返回的内容
type KGTaskRes struct {
	AnswerString string
	AnswerJson   string
	ExecuteTime  int64
	Source       string
	TraceId      string
}

// KGTaskOnceResp 单次请求的所有信息
type KGTaskOnceResp struct {
	Req     *KGTaskReq // 单次测试请求信息
	Res     *KGTaskRes // 单次测试响应信息
	IsPass  bool       // 单次测试测试结果
	EdgCost jsonTime
}

type jsonTime struct {
	time.Duration
}

func (m *KGTask) chanClose() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ChanNum <= 1 {
		close(m.RespChan)
	} else {
		m.ChanNum--
	}
}

// 单条用例执行详情
func (m *KGTask) call(req *KGTaskReq) *KGTaskOnceResp {
	executeTime, _ := strconv.ParseInt(time.Now().Format("20060102150405"), 10, 64)
	Res := &KGTaskOnceResp{
		Req: req,
		Res: &KGTaskRes{
			AnswerString: "",
			AnswerJson:   "",
			ExecuteTime:  executeTime,
		},
	}
	// do test
	startReq := time.Now()

	res := m.TaskConfig.PlatFrom.mChat(m.TaskConfig.Spaces, req.Query)
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
