package models

const (
	Available = 1  // 空闲 switch_on
	Deleted   = 2  // 已删除 switch_off
	Running   = 4  // 执行中
	Stopping  = 16 // 停止中
	Success   = 32 // 执行成功
	Failure   = 64 // 执行失败
)

type List struct {
	PageNum  int `form:"pagenum,default=1" json:"pagenum"`
	PageSize int `form:"pagesize,default=15" json:"pagesize"`
}
