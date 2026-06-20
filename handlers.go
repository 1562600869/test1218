package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, code int, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Code: code, Msg: msg, Data: data})
}

func readBody(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.Unmarshal(body, v)
}

func parseID(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	idStr := parts[len(parts)-1]
	return strconv.ParseInt(idStr, 10, 64)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "static/index.html")
		return
	}
	http.FileServer(http.Dir("static")).ServeHTTP(w, r)
}

func StudentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		list, err := ListStudents()
		if err != nil {
			writeJSON(w, 1, err.Error(), nil)
			return
		}
		writeJSON(w, 0, "success", list)
	case "POST":
		var s Student
		if err := readBody(r, &s); err != nil {
			writeJSON(w, 1, "参数错误: "+err.Error(), nil)
			return
		}
		if s.Name == "" || s.Phone == "" || s.Style == "" {
			writeJSON(w, 1, "姓名、手机、风格必填", nil)
			return
		}
		if s.Level == "" {
			s.Level = "入门"
		}
		id, err := CreateStudent(&s)
		if err != nil {
			writeJSON(w, 1, err.Error(), nil)
			return
		}
		writeJSON(w, 0, "创建成功", map[string]int64{"id": id})
	default:
		http.NotFound(w, r)
	}
}

func PromoteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}

	id, err := parseID(r.URL.Path)
	if err != nil {
		writeJSON(w, 1, "ID错误", nil)
		return
	}

	var req struct {
		Level string `json:"level"`
	}
	if err := readBody(r, &req); err != nil {
		writeJSON(w, 1, "参数错误", nil)
		return
	}

	if err := PromoteStudent(id, req.Level); err != nil {
		writeJSON(w, 1, err.Error(), nil)
		return
	}
	writeJSON(w, 0, "晋升成功", nil)
}

func WorksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		list, err := ListWorks()
		if err != nil {
			writeJSON(w, 1, err.Error(), nil)
			return
		}
		writeJSON(w, 0, "success", list)
	case "POST":
		var ww Work
		if err := readBody(r, &ww); err != nil {
			writeJSON(w, 1, "参数错误: "+err.Error(), nil)
			return
		}
		if ww.StudentID == 0 || ww.WorkName == "" || ww.Subject == "" || ww.CompleteDate == "" {
			writeJSON(w, 1, "学员、作品名、题材、完成日期必填", nil)
			return
		}
		id, err := CreateWork(&ww)
		if err != nil {
			writeJSON(w, 1, err.Error(), nil)
			return
		}
		writeJSON(w, 0, "创建成功", map[string]int64{"id": id})
	default:
		http.NotFound(w, r)
	}
}

func ExhibitionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		list, err := ListExhibitions()
		if err != nil {
			writeJSON(w, 1, err.Error(), nil)
			return
		}
		writeJSON(w, 0, "success", list)
	case "POST":
		var e Exhibition
		if err := readBody(r, &e); err != nil {
			writeJSON(w, 1, "参数错误: "+err.Error(), nil)
			return
		}
		if e.Name == "" || e.StartDate == "" || e.EndDate == "" || e.Location == "" {
			writeJSON(w, 1, "所有字段必填", nil)
			return
		}
		id, err := CreateExhibition(&e)
		if err != nil {
			writeJSON(w, 1, err.Error(), nil)
			return
		}
		writeJSON(w, 0, "创建成功", map[string]int64{"id": id})
	default:
		http.NotFound(w, r)
	}
}

func ExhibitionWorksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	id, err := parseID(r.URL.Path)
	if err != nil {
		writeJSON(w, 1, "ID错误", nil)
		return
	}

	switch r.Method {
	case "GET":
		list, err := ListExhibitionWorks(id)
		if err != nil {
			writeJSON(w, 1, err.Error(), nil)
			return
		}
		writeJSON(w, 0, "success", list)
	case "POST":
		var req struct {
			WorkID int64 `json:"work_id"`
		}
		if err := readBody(r, &req); err != nil {
			writeJSON(w, 1, "参数错误", nil)
			return
		}
		if req.WorkID == 0 {
			writeJSON(w, 1, "作品ID必填", nil)
			return
		}
		if err := AddExhibitionWork(id, req.WorkID); err != nil {
			writeJSON(w, 1, err.Error(), nil)
			return
		}
		writeJSON(w, 0, "参展成功", nil)
	default:
		http.NotFound(w, r)
	}
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}

	list, err := MonthlySubjectStats()
	if err != nil {
		writeJSON(w, 1, err.Error(), nil)
		return
	}
	writeJSON(w, 0, "success", list)
}
