# k8s-app

ローカルの `k8s` 学習用に作る最小構成の Todo アプリです。

## 構成

- Frontend: `TypeScript + React + Vite`
- Backend: `Go`
- デプロイ先: ローカル `k8s`

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

## k8s への適用例

`minikube` を使う場合の一例です。

```bash
minikube addons enable ingress
minikube image load todo-backend:local
minikube image load todo-frontend:local
kubectl apply -f k8s/todo-app.yaml
```

`Ingress NGINX` を有効にした上で、`/etc/hosts` に次を追加します。

```text
127.0.0.1 todo.local
```

その後、`http://todo.local/` にアクセスします。
