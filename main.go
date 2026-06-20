package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "8192", "server port")
	flag.Parse()

	if err := InitDB("./papercut.db"); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer DB.Close()

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/api/students", StudentsHandler)
	http.HandleFunc("/api/students/promote/", PromoteHandler)
	http.HandleFunc("/api/works", WorksHandler)
	http.HandleFunc("/api/exhibitions", ExhibitionsHandler)
	http.HandleFunc("/api/exhibitions/works/", ExhibitionWorksHandler)
	http.HandleFunc("/api/stats/monthly-subjects", StatsHandler)

	addr := ":" + *port
	fmt.Printf("剪纸艺术培训中心管理系统启动成功，访问: http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
