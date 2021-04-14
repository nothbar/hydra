package nfs

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/micro-plat/hydra"
	"github.com/micro-plat/hydra/context"
	"github.com/micro-plat/lib4go/errs"
	"github.com/micro-plat/lib4go/types"
)

//remoting 远程管理
type remoting struct {
	isMaster    bool
	hosts       []string
	masterHost  string
	currentAddr string
}

func newRemoting() *remoting {
	return &remoting{}
}
func (r *remoting) Update(hosts []string, masterHost string, currentAddrs string, isMaster bool) {
	r.hosts = hosts
	r.masterHost = masterHost
	r.currentAddr = currentAddrs
	r.isMaster = isMaster
}

//GetFP 主动向master发起查询,查询某个文件是否存在，及所在的服务器
func (r *remoting) GetFP(name string) (v *eFileFP, err error) {
	//构建请求参数
	v = &eFileFP{}
	input := types.XMap{"name": name}

	//发送远程请求
	log := start(rmt_fp_get, name, r.masterHost)
	rpns, status, err := hydra.C.HTTP().GetRegularClient().Request("POST", fmt.Sprintf("http://%s%s", r.masterHost, rmt_fp_get), input.ToKV(), "utf-8", http.Header{
		context.XRequestID: []string{log.log.GetSessionID()},
		"Accept-Encoding":  []string{"gzip"},
	})

	//处理返回结果
	if status == http.StatusNoContent {
		log.end(rmt_fp_get, name, r.masterHost, status)
		return nil, errs.NewError(http.StatusNotFound, "文件不存在")
	}
	if err != nil {
		log.end(rmt_fp_get, name, r.masterHost, status, err)
		return nil, err
	}

	//处理参数转换
	if err = json.Unmarshal([]byte(rpns), v); err != nil {
		log.end(rmt_fp_get, name, r.masterHost, status, err)
		return nil, err
	}

	//返回结果
	log.end(rmt_fp_get, name, r.masterHost, status)
	return v, nil
}

//Pull 主动从远程服务器拉取文件信息
func (r *remoting) Pull(v *eFileFP) (rpns []byte, err error) {

	//构建请求参数
	input := types.XMap{"name": v.Path}
	host := v.GetAliveHost(r.hosts...)
	if len(host) == 0 {
		return nil, errs.NewError(http.StatusNoContent, "无可用的服务器")
	}

	//向集群发起请求
	var status int
	for _, host := range host {
		log := start(rmt_file_download, v.Path, "from", host)
		rpns, status, err = hydra.C.HTTP().GetRegularClient().Request("POST", fmt.Sprintf("http://%s%s", host, rmt_file_download), input.ToKV(), "utf-8", http.Header{
			context.XRequestID: []string{log.log.GetSessionID()},
			"Accept-Encoding":  []string{"gzip"},
		})

		//检查是否发生错误
		if err != nil {
			log.end(rmt_file_download, v.Path, "from", host, status, err)
			continue
		}

		//检查状态码
		if status == http.StatusNoContent {
			log.end(rmt_file_download, v.Path, "from", host, status)
			continue
		}

		//检查校验位是否一致
		if getCRC64(rpns) != v.CRC64 {
			log.end(rmt_file_download, v.Path, "from", host, status, "crc不一致")
			continue
		}

		//数据正确
		log.end(rmt_file_download, v.Path, "from", host, status)
		return rpns, nil
	}
	return
}

//Report 当前差异时主动向集群推送指纹信息
func (r *remoting) Report(tps eFileFPLists) error {
	//向集群发起请求
	rps := tps.GetAlives(r.getRHosts())
	for host, list := range rps {
		log := start(rmt_fp_notify, host)
		_, status, err := hydra.C.HTTP().GetRegularClient().Request("POST",
			fmt.Sprintf("http://%s%s", host, rmt_fp_notify), types.ToJSON(list), "utf-8", http.Header{
				"Content-Type":     []string{"application/json"},
				context.XRequestID: []string{log.log.GetSessionID()},
				"Accept-Encoding":  []string{"gzip"},
			})
		if err != nil {
			log.end(rmt_fp_notify, host, status, err)
			continue
		}
		log.end(rmt_fp_notify, host, status)
	}
	return nil
}

//Query 向集群机器获取Cache列表,master向所有机器获取,slave启动时向master获取
func (r *remoting) Query() (eFileFPLists, error) {
	//查询数据
	result := make(eFileFPLists)
	for _, host := range r.getQHosts() {

		//查询远程服务
		log := start(rmt_fp_query, "from", host)
		rpns, status, err := hydra.C.HTTP().GetRegularClient().Request("POST", fmt.Sprintf("http://%s%s", host, rmt_fp_query), "", "utf-8", http.Header{
			context.XRequestID: []string{log.log.GetSessionID()},
			"Accept-Encoding":  []string{"gzip"},
		})
		if err != nil {
			log.end(rmt_fp_query, "from", host, status, err)
			return nil, err
		}

		//处理参数合并
		nresult := make(eFileFPLists)
		if err = json.Unmarshal([]byte(rpns), &nresult); err != nil {
			log.end(rmt_fp_query, "from", host, status, err)
			return nil, err
		}

		log.end(rmt_fp_query, "from", host, status)
		for k, v := range nresult {
			if _, ok := result[k]; !ok {
				result[k] = v
				continue
			}
			result[k].MergeHosts(v.Hosts...)
		}
	}
	return result, nil
}

func (r *remoting) getQHosts() []string {
	if !r.isMaster {
		return []string{r.masterHost}
	}
	return r.hosts
}
func (r *remoting) getRHosts() []string {
	if !r.isMaster {
		return []string{r.masterHost}
	}
	mhost := r.hosts
	return append(mhost, r.currentAddr)
}

func fexclude(ex string, hosts ...string) []string {
	mp := make(map[string]interface{})
	nhost := make([]string, 0, len(hosts))
	for _, h := range hosts {
		if _, ok := mp[h]; !ok {
			mp[h] = 0
		}
	}
	for h := range mp {
		if h != ex {
			nhost = append(nhost, h)
		}
	}
	return nhost
}
