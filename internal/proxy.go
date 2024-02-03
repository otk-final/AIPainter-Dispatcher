package internal

import (
	"AIPainter-Dispatcher/internal/ws"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
)

type UserPrincipal struct {
	UserId   string
	UserType string
}

const UserPrincipalKey = "UserPrincipal"

func recodingResponse(recodingHandle func(data []byte)) func(response *http.Response) error {
	return func(response *http.Response) error {
		if recodingHandle == nil {
			return nil
		}

		//解析响应
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		go recodingHandle(body)

		//reset
		response.Body = io.NopCloser(bytes.NewBuffer(body))
		return nil
	}
}

func withProxy(target *url.URL) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL.Scheme = target.Scheme
			request.URL.Host = target.Host
			request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
			request.Host = target.Host
		},
		Transport:     &http.Transport{},
		FlushInterval: 0,
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
			slog.Warn("proxy err")
		},
	}
}

func (m *ComfyUIManager) ApiPrompt(writer http.ResponseWriter, request *http.Request) {
	userPrincipal := request.Context().Value(UserPrincipalKey).(*UserPrincipal)
	userId := userPrincipal.UserId

	//负载均衡，每次请求和目标服务器创建新的长链接
	target := m.lb.Next()

	//查询长链接释放存在
	hold := ws.GetHold(userId)
	if hold != nil {
		hold.NewBind(target)
	}

	proxy := withProxy(target)
	proxy.ModifyResponse = recodingResponse(func(data []byte) {

		//解析响应
		var respJson = map[string]any{}
		_ = json.Unmarshal(data, &respJson)

		//记录任务和目标服务器
		promptId := respJson["prompt_id"]
		_ = m.rdb.HMSet(context.Background(), "PROMPT_ROUTER", promptId, target.String())
	})

	proxy.ServeHTTP(writer, request)
}

type ComfyUIManager struct {
	lb  *LoadBalancer
	rdb *redis.Client
}

const AssetPath = "/Users/hxy"

func NewComfyUIProxy(balancer *LoadBalancer, rdb *redis.Client) *ComfyUIManager {
	return &ComfyUIManager{
		lb:  balancer,
		rdb: rdb,
	}
}

func (m *ComfyUIManager) ApiUpload(writer http.ResponseWriter, request *http.Request) {
	//默认服务器
	withProxy(m.lb.instances[0]).ServeHTTP(writer, request)
}

func (m *ComfyUIManager) uploadLocal(writer http.ResponseWriter, request *http.Request) {
	form := request.MultipartForm
	values := form.Value
	subFolder := values["subfolder"][0]

	//读取文件
	image := form.File["image"][0]
	imagePath := filepath.Join(AssetPath, subFolder, image.Filename)
	f, err := image.Open()
	if err != nil {
		http.Error(writer, "文件解析失败", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	data, _ := io.ReadAll(f)

	//本地存储文件
	_ = os.WriteFile(imagePath, data, os.ModeAppend)
	_ = json.NewEncoder(writer).Encode(map[string]string{"subfolder": subFolder, "filename": image.Filename, "type": "input"})
}

func (m *ComfyUIManager) ApiDownload(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	promptId := vars["prompt_id"]

	if promptId == "" {
		//从本地路径下载
		m.downloadLocal(writer, request)
		return
	}

	//查询任务ID 之前路由到哪个节点
	targetURI, err := m.rdb.HGet(context.Background(), "PROMPT_ROUTER", promptId).Result()
	if errors.Is(err, redis.Nil) {
		http.Error(writer, "not found", http.StatusInternalServerError)
		return
	}

	//查询指定服务节点
	target, _ := url.Parse(targetURI)
	withProxy(target).ServeHTTP(writer, request)
}

func (m *ComfyUIManager) downloadLocal(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	//本地存储文件
	filePath := filepath.Join(AssetPath, vars["subfolder"], vars["filename"])
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(writer, "文件打开失败", http.StatusInternalServerError)
		return
	}

	fileInf, err := os.Stat(filePath)
	if _, err := file.Stat(); os.IsNotExist(err) {
		http.Error(writer, "文件不存在", http.StatusInternalServerError)
		return
	}
	http.ServeContent(writer, request, fileInf.Name(), fileInf.ModTime(), file)
}

func (m *ComfyUIManager) ApiHistory(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	promptId := vars["prompt_id"]

	//查询任务ID 之前路由到哪个节点
	targetURI, err := m.rdb.HGet(context.Background(), "PROMPT_ROUTER", promptId).Result()
	if errors.Is(err, redis.Nil) {
		http.Error(writer, "not found", http.StatusInternalServerError)
		return
	}
	//查询指定服务节点
	target, _ := url.Parse(targetURI)
	withProxy(target).ServeHTTP(writer, request)
}
