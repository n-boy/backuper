# backuper

# pre-build actions
go get -u github.com/jteeuwen/go-bindata/...
go-bindata -pkg base -o base/bindata.go webui/templates/ webui/static/
