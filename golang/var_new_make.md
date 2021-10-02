# go 中 var new make 的区别

`var` 用于声明变量。

`new` 分配内存空间，`func new(Type) *Type` 接收 一个类型，返回这个类型的指针，并将指针指向这个类型的零值（zero value）。

`make` 分配内存空间并根据参数初始化

> 本文主要通过代码示例和原因来解释 var new make 之间的区别。

## var new


通过代码记忆最为合适

[var_new](./var_new_make/var_new/main.go)

## var make slice array

[var_make_slice_array](./var_new_make/var_make_slice_array/doc_test.go)

## var make map

[var_make_map](./var_new_make/var_make_map/doc_test.go)

## make chan

[make_chan](./var_new_make/make_chan/main.go)

