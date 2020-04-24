package main

import (
	"context"
	"fmt"
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
		mux *http.ServeMux
		listener net.Listener
		httpServer *http.Server
		err error
		webServerPort = MasterConf.WebPort
		readTimeOut = time.Duration(MasterConf.ReadTimeOut)
		writeTimeOut = time.Duration(MasterConf.WriteTimeOut)
	)

	mux = http.NewServeMux()
	mux.HandleFunc("/crontab/jobs/list", jobList)
	mux.HandleFunc("/crontab/jobs/post", jobPost)
	mux.HandleFunc("/crontab/jobs/get", jobGet)
	mux.HandleFunc("/crontab/jobs/put", jobPut)
	mux.HandleFunc("/crontab/jobs/delete", jobDelete)

	if listener, err = net.Listen("tcp", ":" + webServerPort); err != nil {
		WebServerInitErr = err
		return
	}

	httpServer = &http.Server{
		ReadTimeout: readTimeOut*time.Millisecond,
		WriteTimeout: writeTimeOut*time.Millisecond,
		Handler: mux,
	}

	go httpServer.Serve(listener)
}

func jobList(w http.ResponseWriter, req *http.Request) {
	var (
		job Job
		err error
		jobs []Job
		jsonResponse = new(JsonResponse)
		rsp []byte
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
		err error
		job = new(Job)
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
		job = new(Job)
		err error
		rsp = new(JsonResponse).NewResponse(0, "")
	)
	if err = req.ParseForm(); err != nil {
		rsp = new(JsonResponse).NewResponse(1, fmt.Sprintf("get request form error: %v", err))
	}
	job.JobInit(req.PostForm.Get("name"), req.PostForm.Get("command"), req.PostForm.Get("cronExpress"))
	if err = job.JobPutHandler(context.TODO()); err != nil {
		rsp = new(JsonResponse).NewResponse(1, err.Error())
	}
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
