.PHONY: all
all:
	GOOS=windows GOARCH=amd64 go build

.PHONY: clean
clean:
	$(RM) dlat.exe
