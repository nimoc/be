# go 中 var new make 的区别

`var` 用于声明变量。

`new` 分配内存空间，`func new(Type) *Type` 接收 一个类型，返回这个类型的指针，并将指针指向这个类型的零值（zero value）。

`make` 分配内存空间并根据参数初始化

> 本文主要通过代码示例和原因来解释 var new make 之间的区别。

## var new

通过代码记忆最为合适

[var\_new](https://github.com/nimoc/be/tree/4f7a513425d41423310b747ec17a13fad2685e59/golang/var_new_make/var_new/main.go)

## var make slice array

[var\_make\_slice\_array](https://github.com/nimoc/be/tree/4f7a513425d41423310b747ec17a13fad2685e59/golang/var_new_make/var_make_slice_array/doc_test.go)

## var make map

[var\_make\_map](https://github.com/nimoc/be/tree/4f7a513425d41423310b747ec17a13fad2685e59/golang/var_new_make/var_make_map/doc_test.go)

## make chan

[make\_chan](https://github.com/nimoc/be/tree/4f7a513425d41423310b747ec17a13fad2685e59/golang/var_new_make/make_chan/main.go)

