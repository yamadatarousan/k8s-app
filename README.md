# k8s-app

ローカルの `k8s` 学習用に作る最小構成の Todo アプリです。

## 構成

- Frontend: `TypeScript + React + Vite`
- Backend: `Go`
- デプロイ先: ローカル `k8s`

## k8s 学習向けに意識している点

- Backend の設定は環境変数で受け取る
- `livenessProbe` 用に `/health` を返す
- `readinessProbe` 用に `/ready` を返す
- ログは標準出力へ出す
- ローカルファイルに状態を持たない
- `SIGTERM` で穏当に止まれるようにする

## ディレクトリ

- `frontend`: ブラウザ向け Todo 画面
- `backend`: Todo API
- `k8s`: Deployment / Service / Ingress

## テスト

```bash
cd backend
go test ./...
```

```bash
cd frontend
npm test
```

## ローカルビルド

```bash
docker build -t todo-backend:local ./backend
docker build -t todo-frontend:local ./frontend
```

## Backend の主な環境変数

- `PORT`: HTTP 待受ポート。既定値は `8080`
- `TODO_API_NAME`: `/health` と `/ready` に載せるアプリ名。既定値は `todo-api`
- `SHUTDOWN_TIMEOUT_SECONDS`: `SIGTERM` 受信後に停止処理へ使う猶予秒数。既定値は `10`

## `kind` での起動手順

実務に近い形を意識して、ローカルクラスターは `kind` を前提にします。

### 1. kind クラスター作成

```bash
kind create cluster --name todo-app --config kind/cluster.yaml
```

### 2. Ingress Controller を導入

`Ingress NGINX` の公式導入手順に合わせて、コントローラーを追加します。

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.15.0/deploy/static/provider/cloud/deploy.yaml
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s
```

### 3. アプリ用イメージをクラスターへ投入

```bash
docker build -t todo-backend:local ./backend
docker build -t todo-frontend:local ./frontend
kind load docker-image todo-backend:local --name todo-app
kind load docker-image todo-frontend:local --name todo-app
```

### 4. マニフェスト適用

```bash
kubectl apply -f k8s/todo-app.yaml
```

### 5. 名前解決を追加

`/etc/hosts` に次を追加します。

```text
127.0.0.1 todo.local
```

その後、`http://todo.local/` にアクセスします。

## 補足

- `kind` のクラスタ定義は [kind/cluster.yaml](/Users/user/Development/k8s-app/kind/cluster.yaml) にあります。
- `Ingress NGINX` の導入元は公式ドキュメントです。
  https://kubernetes.github.io/ingress-nginx/deploy/
- `kind` で Ingress を扱う前提は公式ドキュメントの構成に合わせています。
  https://kind.sigs.k8s.io/docs/user/ingress/
- 2026年3月13日時点では `Ingress NGINX` は引退告知済みで、2026年3月まで best-effort maintenance と案内されています。学習用にはまだ有用ですが、中長期では `Gateway API` も合わせて学ぶ前提で捉えるのが妥当です。
