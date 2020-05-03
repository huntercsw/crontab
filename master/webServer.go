package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"net"
	"net/http"
	"sync"
	"time"
)

var WebServerInitErr error
var webInitOnce sync.Once

func WebServerInit() {
	webInitOnce.Do(webServerInitOnce)
}

func webServerInitOnce() {
	var (
		mux           *http.ServeMux
		listener      net.Listener
		httpServer    *http.Server
		err           error
		webServerPort = MasterConf.WebPort
		readTimeOut   = time.Duration(MasterConf.ReadTimeOut)
		writeTimeOut  = time.Duration(MasterConf.WriteTimeOut)
	)

	mux = http.NewServeMux()
	mux.HandleFunc("/crontab/jobs/list", jobList)
	mux.HandleFunc("/crontab/jobs/post", jobPost)
	mux.HandleFunc("/crontab/jobs/get", jobGet)
	mux.HandleFunc("/crontab/jobs/put", jobPut)
	mux.HandleFunc("/crontab/jobs/delete", jobDelete)
	mux.HandleFunc("/crontab/jobs/kill", jobKill)

	if listener, err = net.Listen("tcp", ":"+webServerPort); err != nil {
		WebServerInitErr = err
		return
	}

	httpServer = &http.Server{
		ReadTimeout:  readTimeOut * time.Millisecond,
		WriteTimeout: writeTimeOut * time.Millisecond,
		Handler:      mux,
	}

	go httpServer.Serve(listener)
}

func jobList(w http.ResponseWriter, req *http.Request) {
	var (
		job          Job
		err          error
		jobs         []Job
		jsonResponse = new(JsonResponse)
		rsp          []byte
	)
	if jobs, err = job.JobListHandler(context.TODO()); err != nil {
		rsp = jsonResponse.NewResponse(1, fmt.Sprintf("get jobs list error"))
	} else {
		rsp = jsonResponse.NewResponse(0, jobs)
	}
	w.Write(rsp)
}

func jobPost(w http.ResponseWriter, req *http.Request) {
	var (
		err          error
		job          = new(Job)
		jsonResponse JsonResponse
	)
	rsp := jsonResponse.NewResponse(0, "")
	if err = req.ParseForm(); err != nil {
		rsp = jsonResponse.NewResponse(1, fmt.Sprintf("get request form error: %v", err))
	}
	job.JobInit(req.PostForm.Get("name"), req.PostForm.Get("command"), req.PostForm.Get("cronExpress"))
	if err = job.JobPostHandler(context.TODO()); err != nil {
		msg := fmt.Sprintf("add job to etcd error: %v", err)
		MasterLogger.Error.Println(msg)
		rsp = jsonResponse.NewResponse(1, msg)
	}
	w.Write(rsp)
}

func jobGet(w http.ResponseWriter, req *http.Request) {
}

func jobPut(w http.ResponseWriter, req *http.Request) {
	var (
		job           = new(Job)
		err           error
		rsp           = new(JsonResponse).NewResponse(0, "")
		originJobName string
	)

	if req.URL.Query().Get("jobName") == "" {
		rsp = new(JsonResponse).NewResponse(1, "job name is required")
		goto RESPONSE
	}

	originJobName = JOB_PATH + req.URL.Query().Get("jobName")

	if err = req.ParseForm(); err != nil {
		rsp = new(JsonResponse).NewResponse(1, fmt.Sprintf("get request form error: %v", err))
		goto RESPONSE
	}

	job.JobInit(req.PostForm.Get("name"), req.PostForm.Get("command"), req.PostForm.Get("cronExpress"))
	if err = job.JobPutHandler(context.TODO(), originJobName); err != nil {
		rsp = new(JsonResponse).NewResponse(1, err.Error())
	}

RESPONSE:
	w.Write(rsp)
}

func jobDelete(w http.ResponseWriter, req *http.Request) {
	var (
		job = new(Job)
		err error
		rsp = new(JsonResponse).NewResponse(0, "")
	)
	job.Name = req.URL.Query().Get("jobName")
	if err = job.JobDeleteHandler(context.TODO()); err != nil {
		rsp = new(JsonResponse).NewResponse(1, err.Error())
	}
	w.Write(rsp)
}

func jobKill(w http.ResponseWriter, req *http.Request) {
	var (
		err   error
		job   = new(Job)
		count int64
		kvs   []*mvccpb.KeyValue
		rsp   = new(JsonResponse).NewResponse(0, "")
	)
	job.JobInit(req.PostForm.Get("name"), req.PostForm.Get("command"), req.PostForm.Get("cronExpress"))

	if count, kvs, err = job.JobGetHandler(context.TODO()); err != nil {
		rsp = new(JsonResponse).NewResponse(1, err.Error())
	}

	if count == 0 {
		rsp = new(JsonResponse).NewResponse(1, fmt.Sprintf("job[%s] dose not exist", job.Name))
	}

	k, v := kvs[0].Key, kvs[0].Value
	jobDest := new(Job)
	if err = json.Unmarshal(v, jobDest); err != nil {
		MasterLogger.Error.Println(fmt.Sprintf("josn.Unmarshal job[%s] error: %v", string(k), err))
		rsp = new(JsonResponse).NewResponse(1, "json Unmarshal error")
	} else {
		if jobDest.Status != JOB_STATUS_RUNNING {
			rsp = new(JsonResponse).NewResponse(1, fmt.Sprintf("job[%s] is not running", jobDest.Name))
		} else {
			if err1 := jobDest.JobKillHandler(context.TODO()); err1 != nil {
				rsp = new(JsonResponse).NewResponse(1, err1.Error())
			}
		}
	}
	w.Write(rsp)
}
