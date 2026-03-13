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

`kind` では `provider/cloud` ではなく `provider/kind` を使います。
ここを間違えると Controller Pod は起動しても `http://todo.local/` に到達しません。

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.15.0/deploy/static/provider/kind/deploy.yaml
kubectl -n ingress-nginx patch deployment ingress-nginx-controller --type='merge' -p '{"spec":{"template":{"spec":{"nodeSelector":{"kubernetes.io/os":"linux","ingress-ready":"true"}}}}}'
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

### 6. 到達確認

ブラウザを開く前に、`todo.local` へ到達できることを確認します。

```bash
kubectl -n ingress-nginx get pods
kubectl get ingress
curl -I -H 'Host: todo.local' http://127.0.0.1/
```

`HTTP/1.1 200 OK` か `HTTP/1.1 304 Not Modified` が返れば、`http://todo.local/` へアクセスできます。

## `kind` クラスターの停止・再開・再作成

ローカル開発では「一時停止」と「完全削除」を分けて扱います。

### 一時停止（状態を残す）

`Pod` や `Service` を残したまま、`kind` ノードコンテナだけ停止します。

```bash
docker stop $(docker ps -a --filter "label=io.x-k8s.kind.cluster=todo-app" --format '{{.Names}}')
```

### 再開（停止した状態から戻す）

停止したノードコンテナを起動し、`kubectl` の接続先を戻します。

```bash
docker start $(docker ps -a --filter "label=io.x-k8s.kind.cluster=todo-app" --format '{{.Names}}')
kubectl config use-context kind-todo-app
```

### 完全削除（作り直し）

`delete` は破壊的操作です。クラスター内のリソースは消えるため、その後は本 README の「起動手順」を 1 から再実行する必要があります。

```bash
kind delete cluster --name todo-app
kind create cluster --name todo-app --config kind/cluster.yaml
```

## 補足

- `kind` のクラスタ定義は [kind/cluster.yaml](/Users/user/Development/k8s-app/kind/cluster.yaml) にあります。
- `Ingress NGINX` の導入元は公式ドキュメントです。
  https://kubernetes.github.io/ingress-nginx/deploy/
- `kind` で Ingress を扱う前提は公式ドキュメントの構成に合わせています。
  https://kind.sigs.k8s.io/docs/user/ingress/
- 2026年3月13日時点では `Ingress NGINX` は引退告知済みで、2026年3月まで best-effort maintenance と案内されています。学習用にはまだ有用ですが、中長期では `Gateway API` も合わせて学ぶ前提で捉えるのが妥当です。
