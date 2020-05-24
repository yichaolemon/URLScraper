RUN = docker run --rm -v $(CURDIR):/usr/src/URLScraper -w /usr/src/URLScraper golang:1.13-alpine

URLScraper: main.go
	$(RUN) go build -v

run: URLScraper
	$(RUN) ./URLScraper