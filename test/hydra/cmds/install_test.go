package cmds

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/micro-plat/hydra"
	"github.com/micro-plat/hydra/hydra/servers/http"
	"github.com/micro-plat/hydra/test/assert"
)

const installRegistryAddr = "lm://."

func Test_install_Normal(t *testing.T) {
	resetServiceName(t.Name())
	execPrint(t)
	defunc, fileCallback := injectStdOutFile()
	defer defunc()

	var app = hydra.NewApp(
		hydra.WithServerTypes(http.API),
		hydra.WithPlatName("xxtest"),
		hydra.WithSystemName("apiserver"),
		hydra.WithClusterName("c"),
	)

	//2. 清除服务(保证没有服务安装)
	os.Args = []string{"xxtest", "remove"}
	go app.Start()
	time.Sleep(time.Second * 2)

	//正常的安装
	os.Args = []string{"xxtest", "install", "-r", installRegistryAddr, "-c", "c"}
	app.Start()
	time.Sleep(time.Second)

	//3. 清除服务
	os.Args = []string{"xxtest", "remove"}
	go app.Start()

	time.Sleep(time.Second * 2)
	bytes, err := fileCallback()

	if err != nil {
		t.Error(err)
		return
	}
	lines := strings.Split(string(bytes), "\r")
	for _, row := range lines {
		if !strings.Contains(row, "Install") {
			continue
		}
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			//unbuntu/centos
			result := strings.Contains(row, "sudo") || strings.Contains(row, "OK")
			assert.Equal(t, true, result, "正常参数的安装")
		}
		if runtime.GOOS == "windows" {
			result := strings.Contains(row, "OK")
			assert.Equal(t, true, result, "正常参数的安装")
		}
		return
	}
	t.Error("未找到安装的输出信息")
}

func Test_install_Less_param(t *testing.T) {
	resetServiceName(t.Name())
	execPrint(t)

	defunc, fileCallback := injectStdOutFile()
	defer defunc()

	var app = hydra.NewApp(
		hydra.WithServerTypes(http.API),
		hydra.WithPlatName("xxtest"),
		hydra.WithSystemName("apiserver"),
		//hydra.WithClusterName("c"),
	)

	//2. 清除服务(保证没有服务安装)
	os.Args = []string{"xxtest", "remove"}
	go app.Start()
	time.Sleep(time.Second * 2)

	//缺少参数的安装 -c
	os.Args = []string{"xxtest", "install", "-r", installRegistryAddr}
	app.Start()
	time.Sleep(time.Second)

	//2. 删除服务
	os.Args = []string{"xxtest", "remove"}
	app.Start()

	time.Sleep(time.Second * 2)
	bytes, err := fileCallback()

	if err != nil {
		t.Error(err)
		return
	}
	lines := strings.Split(string(bytes), "\r")
	for _, row := range lines {
		if !strings.Contains(row, "Install") {
			continue
		}
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			//unbuntu/centos
			result := strings.Contains(row, "sudo") || strings.Contains(row, "集群名不能为空称")
			assert.Equal(t, true, result, "缺少参数的安装 -c")
		}
		if runtime.GOOS == "windows" {
			result := strings.Contains(row, "集群名不能为空称")
			assert.Equal(t, true, result, "缺少参数的安装 -c")
		}
		return
	}
}

func Test_install_Cover(t *testing.T) {
	resetServiceName(t.Name())
	execPrint(t)

	defunc, fileCallback := injectStdOutFile()
	defer defunc()

	//覆盖安装 -c
	args := []string{"xxtest", "install", "-r", installRegistryAddr, "-c", "c", "-cover", "true"}

	var app = hydra.NewApp(
		hydra.WithServerTypes(http.API),
		hydra.WithPlatName("xxtest"),
		hydra.WithSystemName("apiserver"),
		hydra.WithClusterName("c"),
	)
	os.Args = args
	app.Start()

	time.Sleep(time.Second * 2)
	bytes, err := fileCallback()

	if err != nil {
		t.Error(err)
		return
	}
	lines := strings.Split(string(bytes), "\r")
	for _, row := range lines {
		if !strings.Contains(row, "Install") {
			continue
		}
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			//unbuntu/centos
			result := strings.Contains(row, "sudo") || strings.Contains(row, "OK")
			assert.Equal(t, true, result, "覆盖安装 -cover=true")
		}
		if runtime.GOOS == "windows" {
			result := strings.Contains(row, "OK")
			assert.Equal(t, true, result, "覆盖安装 -cover=true")
		}
		return
	}
}