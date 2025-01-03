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
go install github.com/ieee0824/ndcd-go/ndcd -i .github/before.jpg -o .github/after.jpg -oh 64 -bt box -bs 10 -c 1 -g 0.8 -s -oe 512
```

## sample

<table border="0">
<tr>
<td><img height="256px" src=".github/before.jpg">
<td><img height="256px" src=".github/after.jpg">
</tr>
</table>




