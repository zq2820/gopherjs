module github.com/gopherjs/gopherjs

go 1.18

require (
	github.com/fsnotify/fsnotify v1.5.4
	github.com/google/go-cmp v0.5.8
	github.com/gorilla/websocket v1.5.0
	github.com/neelance/astrewrite v0.0.0-20160511093645-99348263ae86
	github.com/neelance/sourcemap v0.0.0-20200213170602-2833bce08e4c
	github.com/shurcooL/go v0.0.0-20200502201357-93f07166e636
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/tdewolff/minify/v2 v2.11.11
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f
	golang.org/x/sys v0.0.0-20220624220833-87e55d714810
	golang.org/x/tools v0.1.11
	honnef.co/go/js/console v0.0.0-20150119023344-105276c43558
	honnef.co/go/js/dom v0.0.0-20210725211120-f030747120f2
	myitcv.io v0.0.0-20201125173645-a7167afc9e13
)

require (
	github.com/gopherjs/jsbuiltin v0.0.0-20180426082241-50091555e127 // indirect
	github.com/tdewolff/parse/v2 v2.6.0 // indirect
)

require (
	github.com/gorilla/css v1.0.0
	github.com/jinzhu/copier v0.3.5
	github.com/speps/go-hashids/v2 v2.0.1
	github.com/vanng822/css v1.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/speps/go-hashids v1.0.0
	github.com/wellington/go-libsass v0.9.2
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2
	golang.org/x/term v0.0.0-20220526004731-065cf7ba2467 // indirect
)

replace github.com/gopherjs/gopherjs v1.17.2 => ./

replace myitcv.io v0.0.0-20201125173645-a7167afc9e13 => github.com/zq2820/x v0.0.0-20220723033027-f460afcd9569