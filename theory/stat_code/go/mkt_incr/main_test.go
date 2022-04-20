package main_test

import (
	"log"
	"testing"
)

// 调用空结构体指针的方法
type A struct {}
func (A) Hi() {log.Print(1)}
func Test_1(t *testing.T) {
	var a *A
	a.Hi()
}

// 调用结构体未赋值的函数
type B struct {
	Hi func()
}
func Test_2(t *testing.T) {
	var b B
	b.Hi()
}

// 调用空接口的方法
type C interface {
	Hi()
}
func Test_3(t *testing.T) {
	var c C
	c.Hi()
}
