package smartvoice

import (
	"context"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	talk "task-go/apis/proto/smartvoice"
	"task-go/pkg/app"
	"task-go/pkg/e"
	"task-go/pkg/logf"
	util "task-go/pkg/util/const"
	"time"
)

func smartVoiceCall(conn *grpc.ClientConn, r *talk.TalkRequest) (resp *talk.TalkResponse, duration time.Duration, err error) {
	c := talk.NewTalkClient(conn)
	startTime := time.Now()
	resp, err = c.Talk(context.Background(), r)
	if err != nil {
		return nil, 0, err
	}
	duration = time.Now().Sub(startTime)
	return resp, duration, err
}

func smartVoiceDail(addr string) (conn *grpc.ClientConn, err error) {
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, err
}

func SmartVoiceTestOne(context *gin.Context) {
	//BindAndValid
	copyReq := &talk.TalkRequest{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error())
		logf.Debug("Validation", err.Error())
		context.Abort()
	}
	context.Set(util.REQUEST_KEY, copyReq)

	//VerifyReq
	r := context.MustGet(util.REQUEST_KEY).(*talk.TalkRequest)

	conn, err := smartVoiceDail(context.Param("addr"))
	if err != nil {
		app.ErrorResp(context, 500, "dial grpc addr err! "+err.Error())
	} else {
		defer func(conn *grpc.ClientConn) {
			err := conn.Close()
			if err != nil {
				app.ErrorResp(context, 500, "close grpc conn err! "+err.Error())
			}
		}(conn)
		resp, duration, err := smartVoiceCall(conn, r)
		if err != nil {
			app.ErrorResp(context, 500, "smartVoiceCall error! "+err.Error())
		} else {
			app.SuccessResp(context, struct {
				Cost int64       `json:"cost"`
				Data interface{} `json:"data"`
			}{
				Cost: duration.Milliseconds(),
				Data: resp,
			})
		}
	}
}
