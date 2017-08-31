gcal2tsv-go
===

# 事前準備

## バイナリダウンロード

`./bin/` から各環境に合わせたものを、適宜取得。

## ビルドする場合

```
$ go get -u github.com/terukizm/gcal2tsv-go
$ go get -u google.golang.org/api/calendar/v3
$ go get -u golang.org/x/oauth2/...
```

```
# win(32bit/64bit)
$ GOOS=windows GOARCH=386 go build -o ./bin/windows386/gcal2tsv.exe
$ GOOS=windows GOARCH=amd64 go build -o ./bin/windows64/gcal2tsv.exe

# mac(32bit/64bit)
$ GOOS=darwin GOARCH=386 go build -o ./bin/darwin386/gcal2tsv
$ GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin64/gcal2tsv

もしくは

$ go run gcal2tsv.go
```

# 使い方

1. [Step 1: Turn on the Google Calendar API](https://developers.google.com/google-apps/calendar/quickstart/go)を参考に`client_secret.json`を取得
2. `config.toml.sample`を`config.toml`にリネーム
3. `config.toml`を適宜設定
    * 集計開始日、集計終了日
    * calendar_idに出力対象となるGoogleカレンダーのIDを設定(カレンダーの情報->カレンダーのアドレス->カレンダーID で確認可能)
4. 各環境用のバイナリ(例: `gcal2tsv.exe`)と`config.toml`を同じディレクトリに配置、`client_secret.json`のパスを`config.toml`で設定したものに合わせる
4. `gcal2tsv.exe`を実行
    * 初回はWebブラウザを利用した認証が行われる。認証情報(credential)は ~/.credentials/ に格納される

## 参考

```
terukizm-MBP13:gcal2tsv-go terukizm$ mkdir ~/gcal2tsv
terukizm-MBP13:gcal2tsv-go terukizm$ cp bin/darwin64/gcal2tsv ~/gcal2tsv/
terukizm-MBP13:gcal2tsv-go terukizm$ cp config.toml ~/gcal2tsv/
terukizm-MBP13:gcal2tsv-go terukizm$ cp client_secret.json ~/gcal2tsv/

terukizm-MBP13:gcal2tsv-go terukizm$ cd ~/gcal2tsv/
terukizm-MBP13:gcal2tsv terukizm$ ls
client_secret.json	config.toml		gcal2tsv

terukizm-MBP13:gcal2tsv terukizm$ ./gcal2tsv
start=2017-08-01, end=2017-08-31
[2017-08-30 21:00:00 2017-08-30 22:30:00 quickstart.goをベースに実装開始 1.5]
[2017-08-30 23:00:00 2017-08-31 00:45:00 githubに上げたりPR作ったりエディタいじったりとか 1.75]
```