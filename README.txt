1.タイトル
RSS Aggregator

2.概要
Goのバックエンド開発を学習するために作成したRSS Aggregatorです。
ユーザーがRSSフィードを登録・フォローし、
バックグラウンドでRSSから記事を取得してDBへ保存できます。
また、ユーザーは自分がフォローしているフィードの記事一覧をAPI経由で取得できます。
YouTubeのGo学習教材  
「Go Programming – Golang Course with Bonus Projects」
（動画リンク："https://youtu.be/un6ZyFkqFKo"）
を参考にしながら、実際にコードを書き、エラーを調査し、各処理の意味を確認しながら実装しました。

3.学習テーマ
・HTTPサーバ、ルーティング
・REST API
・PostgreSQLとの接続
・マイグレーション
・goroutineを使った並行処理、WaitGroupによるgoroutineの待機
・定期実行

4.主な機能
・ユーザー登録
・APIキーによるユーザー認証
・RSSフィード登録
・RSSフィード一覧取得
・フィードのフォロー
・フィードのフォロー解除
・フォロー中フィードの取得
・RSS記事の定期取得
・RSS記事のPostgreSQLへの保存
・フォローしているフィードの記事一覧取得

5.使用技術
・Go
・PostgreSQL
・chi
・database/sql
・goose
・sqlc

8.学んだこと
変数や関数、構造体、ポインタなどの基礎文法を学習し、Webバックエンドのなかでどのように組み合わせるのかを学びました。
また、特に理解が深まったのは以下2点です。
①HTTPリクエスト⇒Router⇒handlerへ渡される流れとその書き方
②sqlcによってSQLから型安全なGoコードを生成し、DB操作をGoから実現する方法


