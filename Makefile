build:
	env GOOS=darwin  GOARCH=amd64 go build -o bin/hipchat-gerrit.darwin main.go
	env GOOS=linux   GOARCH=amd64 go build -o bin/hipchat-gerrit main.go
	env GOOS=windows GOARCH=amd64 go build -o bin/hipchat-gerrit.exe    main.go
