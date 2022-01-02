# go-sync-tcp
C/S架构的分布式文件传输程序

### use help

切换到代码路径下后如下述方式进行运行

```bash
go run master.go /gopath/src/dis_test_file/master/test_file/1024k.txt 3000 > /gopath/src/dis_test_file/master/result/size5.txt



go run master_unlock.go /gopath/src/dis_test_file/master/test_file/4k.txt 3000

go run master_gj.go /gopath/src/dis_test_file/master/test_file/4k.txt 9000



go run master_gj.go /gopath/src/dis_test_file/master/test_file/4k.txt 30000

```

