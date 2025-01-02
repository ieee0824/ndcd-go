# ndcd go
画像をドット画風に変換します

## アルゴリズム
[こちら](https://www.jstage.jst.go.jp/article/itej/74/3/74_597/_pdf)の論文を参考にしています


## install command

```
go install github.com/ieee0824/ndcd-go/ndcd
```

## command
```
go run ndcd/main.go -i .github/before.jpg -o .github/after.jpg -oh 32
```

## sample
<img height="64px" src=".github/before.jpg">
<img height="64px" src=".github/after.jpg">