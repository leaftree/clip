package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

var (
	fileSourceRoot string

	ErrOK                  = NewE(0, "成功")
	ErrFilePathInvalid     = NewE(1, "文件路径不合法")
	ErrNotSupportedType    = NewE(2, "不支持响应类型")
	ErrServerInternalError = NewE(3, "服务器内部异常")
)

func init() {
	fileSourceRoot = filepath.Join(os.Getenv("HOME"), ".clip/source")
}

func main() {
	host := GetAddr()

	s := NewServe(host)
	s.Register("/u/{file}", s.handler, http.MethodGet)

	log.Println("start clip service on: ", host)
	log.Fatal(s.Serve())
}

type serve struct {
	host   string
	router *mux.Router
}

func NewServe(host string) *serve {
	return &serve{
		host:   host,
		router: mux.NewRouter(),
	}
}

func (s *serve) Serve() error {
	return http.ListenAndServe(s.host, s.router)
}

func (s *serve) Register(path string, hdr http.HandlerFunc, methods ...string) *serve {
	s.router.HandleFunc(path, hdr).Methods(methods...)
	return s
}

type Handler struct {
	ctx         context.Context
	w           http.ResponseWriter
	r           *http.Request
	contentType string
}

func NewHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) *Handler {
	h := Handler{ctx: ctx, w: w, r: r}
	h.contentType = "application/json"
	return &h
}

func (h *Handler) SetResponseType(contentType string) *Handler {
	h.contentType = contentType
	return h
}

func (h *Handler) Write(val interface{}) error {
	var err error
	var data []byte

	switch h.contentType {
	case "application/type":
		if data, err = json.Marshal(val); err != nil {
			return err
		}

	case "plain/text":
		data = val.([]byte)

	default:
		h.w.WriteHeader(500)
		h.w.Write([]byte("server error"))
		return ErrNotSupportedType
	}

	h.w.WriteHeader(200)
	h.w.Write(data)
	return nil
}

func (h *Handler) Writefile(reader io.Reader) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	h.w.WriteHeader(200)
	h.w.Write(data)
	return nil
}

func (h *Handler) Error(ec *E) {
	if ec == nil {
		ec = ErrOK
	}
	data, _ := json.Marshal(map[string]interface{}{
		"msg":  ec.msg,
		"code": ec.code,
	})
	h.w.Write(data)
}

type E struct {
	msg  string
	code uint32
}

func NewE(code uint32, msg string) *E {
	return &E{msg: msg, code: code}
}

func (e E) Error() string {
	return fmt.Sprintf("error(%d): %s", e.code, e.msg)
}

func (s *serve) handler(w http.ResponseWriter, r *http.Request) {
	hdr := NewHandler(context.Background(), w, r)

	filename := filepath.Join(fileSourceRoot, mux.Vars(r)["file"])

	if filename == fileSourceRoot {
		hdr.Error(ErrFilePathInvalid)
		return
	}

	fd, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		hdr.Error(ErrServerInternalError)
		return
	}
	defer fd.Close()

	err = hdr.Writefile(fd)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("response file ", filename)
}
