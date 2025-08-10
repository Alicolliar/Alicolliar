build-server:
	env GOOS=linux GOARCH=amd64 go build -o=./build/blog-platform-server -ldflags="-X 'main.SITE_BASEURL=https://alicolliar.net'" .

build-posting-tui:
	cd blogPostingApp; env GOOS=linux GOARCH=amd64 && go build -o=../build/blog-tui  .
