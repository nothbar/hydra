package services

import (
	"fmt"

	"github.com/micro-plat/hydra/context"
	"github.com/micro-plat/hydra/registry/conf/server"
)

type serverHook struct {
	starting  func(server.IServerConf) error
	closing   func(server.IServerConf) error
	handlings []context.IHandler
	handleds  []context.IHandler
}

//AddStarting 添加服务器启动勾子
func (s *serverHook) AddStarting(h func(server.IServerConf) error) error {
	if h == nil {
		return fmt.Errorf("启动服务不能为空")
	}
	if s.starting != nil {
		return fmt.Errorf("启动服务不能重复注册")
	}
	s.starting = h
	return nil
}

//AddClosing 设置关闭器关闭勾子
func (s *serverHook) AddClosing(h func(server.IServerConf) error) error {
	if h == nil {
		return fmt.Errorf("关闭服务不能为空")
	}
	if s.closing != nil {
		return fmt.Errorf("启动服务不能重复注册")
	}
	s.closing = h
	return nil
}

//AddHandleExecuting 添加handle预处理勾子
func (s *serverHook) AddHandleExecuting(h ...context.IHandler) error {
	if len(h) == 0 {
		return nil
	}
	s.handlings = append(s.handlings, h...)
	return nil
}

//AddHandleExecuted 添加handle后处理勾子
func (s *serverHook) AddHandleExecuted(h ...context.IHandler) error {
	if len(h) == 0 {
		return nil
	}
	s.handleds = append(s.handleds, h...)
	return nil
}

//GetHandleExecutings 获取handle预处理勾子
func (s *serverHook) GetHandleExecutings() []context.IHandler {
	return s.handlings
}

//GetHandleExecuteds 获取handle后处理勾子
func (s *serverHook) GetHandleExecuteds() []context.IHandler {
	return s.handleds
}

//OnStarting 获取服务器启动预处理函数
func (s *serverHook) OnStarting(c server.IServerConf) error {
	if s.starting == nil {
		return nil
	}
	return s.starting(c)
}

//GetClosings 获取服务器关闭处理函数
func (s *serverHook) OnClosing(c server.IServerConf) error {
	if s.closing == nil {
		return nil
	}
	return s.closing(c)
}
