FROM golang:1.25-bookworm

ARG LIQUIBASE_VERSION=4.33.0
ARG POSTGRES_JDBC_VERSION=42.7.7

WORKDIR /app

# Install Java and basic tools required by Liquibase.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        openjdk-21-jre-headless \
        tar \
    && rm -rf /var/lib/apt/lists/*

# Install Liquibase CLI.
RUN mkdir -p /opt/liquibase \
    && curl -fsSL "https://github.com/liquibase/liquibase/releases/download/v${LIQUIBASE_VERSION}/liquibase-${LIQUIBASE_VERSION}.tar.gz" \
        | tar -xz -C /opt/liquibase \
    && chmod +x /opt/liquibase/liquibase \
    && ln -s /opt/liquibase/liquibase /usr/local/bin/liquibase

# Liquibase includes the PostgreSQL driver in recent releases, but keep one explicit
# driver in the image so migrations work even if the bundled driver changes later.
RUN mkdir -p /opt/liquibase/lib \
    && curl -fsSL \
        "https://jdbc.postgresql.org/download/postgresql-${POSTGRES_JDBC_VERSION}.jar" \
        -o /opt/liquibase/lib/postgresql.jar

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Railway runs this script, and it runs Liquibase before starting the Go backend.
RUN chmod +x scripts/*.sh \
    && go build -o /tmp/potential-customer-agency-api ./cmd/api

EXPOSE 8080

CMD ["./scripts/railway-start.sh"]
