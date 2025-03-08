compose_postgres_dsn() {
  local host="${POSTGRES_HOST:-localhost}"
  local port="${POSTGRES_PORT:-5432}"
  local user="${POSTGRES_USER:-knowledge_leaf}"
  local password="${POSTGRES_PASSWORD}"
  local dbname="${POSTGRES_DATABASE:-knowledge_leaf}"
  local sslmode="${POSTGRES_SSLMODE:-disable}" # default to disable ssl

  local dsn="postgresql://"

  if [[ -n "$user" ]]; then
    dsn+="$user"
    if [[ -n "$password" ]]; then
      dsn+=":$password"
    fi
    dsn+="@"
  fi

  dsn+="$host"

  if [[ -n "$port" ]]; then
    dsn+=":$port"
  fi

  dsn+="/$dbname"

  if [[ -n "$sslmode" ]]; then
    dsn+="?sslmode=$sslmode"
  fi

  echo "$dsn"
}

if type -P migrate >/dev/null 2>&1; then
  echo 'migrate executable found'
else
  echo 'installing migrate CLI via Go toolchain'
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

export GOBIN="$(go env GOPATH)/bin"
$GOBIN/migrate -database "$(compose_postgres_dsn)" -source 'file://migrations/postgres' up

