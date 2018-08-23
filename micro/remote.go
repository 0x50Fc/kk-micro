package micro

import (
	"encoding/json"
	"errors"
	"log"
	"sync"

	"github.com/hailongz/kk-lib/dynamic"
	"github.com/hailongz/kk-lib/http"
)

type IRemote interface {
	Handle(task ITask) error
}

type HttpRemote struct {
	baseURL string
}

func NewHttpRemote(baseURL string) *HttpRemote {
	v := HttpRemote{}
	v.baseURL = baseURL
	return &v
}

func (R *HttpRemote) Handle(task ITask) error {

	options := http.Options{}
	options.Type = http.OptionTypeJson
	options.ResponseType = http.OptionResponseTypeByte
	options.Method = "POST"
	options.Data = task
	options.Url = R.baseURL + task.GetName() + ".json"

	log.Println("[HTTP]", options.Url)

	data, err := http.Send(&options)

	if err != nil {
		return err
	}

	log.Println("[HTTP]", string(data.([]byte)))

	r := task.GetResult()

	if r != nil && data != nil {
		err = json.Unmarshal(data.([]byte), r)
		if err != nil {
			return err
		}
	}

	return nil
}

type RemoteTask struct {
	Remote string
	Task   ITask
}

func (T *RemoteTask) GetName() string {
	return T.Task.GetName()
}

func (T *RemoteTask) GetTitle() string {
	return T.Task.GetTitle()
}

func (T *RemoteTask) GetResult() interface{} {
	return T.Task.GetResult()
}

type RemoteService struct {
	lock    sync.Mutex
	remotes map[string]IRemote
}

func (S *RemoteService) GetTitle() string {
	return "Remote 服务"
}

func (S *RemoteService) HandleRemoteTask(app IApp, task *RemoteTask) error {

	S.lock.Lock()

	if S.remotes == nil {
		S.remotes = map[string]IRemote{}
	}

	remote, ok := S.remotes[task.Remote]

	S.lock.Unlock()

	if !ok {

		v := dynamic.Get(app.Config(), task.Remote)

		if v == nil {
			return errors.New("未找到远程服务配置 " + task.Remote)
		}

		stype := dynamic.StringValue(dynamic.Get(v, "type"), "http")

		if stype == "http" {
			remote = NewHttpRemote(dynamic.StringValue(dynamic.Get(v, "baseURL"), ""))
		} else {
			return errors.New("不支持的远程服务类型 " + stype)
		}

		S.lock.Lock()

		S.remotes[task.Remote] = remote

		S.lock.Unlock()

	}

	return remote.Handle(task.Task)

}

func RemoteHandle(app IApp, remote string, task ITask) error {
	v := RemoteTask{}
	v.Remote = remote
	v.Task = task
	return app.Handle(&v)
}
