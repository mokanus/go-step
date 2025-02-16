@SET GOPATH=C:\Work\crazyant\crazyant-codes\lbx-server
@SET CGO_ENABLED=0

cd..
go build -ldflags -w -o .\protoc-gen-go\protoc-gen-go.exe .\protoc-gen-go
pause