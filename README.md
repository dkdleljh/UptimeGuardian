# UptimeGuardian

Go-based health check and alert bot for server endpoints.

## 한국어 (Korean)

UptimeGuardian은 서버/웹서비스 상태를 주기적으로 확인하고, 상태 변화(장애/복구) 시 웹훅으로 알림을 보내는 프로그램입니다.

### 주요 기능

- 체크 대상 CRUD API
- 수동 실행 API (`POST /checks/{id}/run`)
- 체크별 주기 실행 (`intervalSec`)
- 상태 변화 알림 (DOWN, 선택적으로 RECOVERY)
- JSON 파일 기반 영속 저장

### 빠른 시작

```bash
cp .env.example .env
docker compose up -d --build
curl http://localhost:8080/health
```

### 환경 변수

- `PORT` (기본 `8080`)
- `DATA_FILE` (기본 `./data/checks.json`)
- `ALERT_WEBHOOK_URL` (비우면 알림 비활성)
- `REQUEST_USER_AGENT`
- `API_KEY` (선택) 설정 시 API 요청에 `X-API-Key` 또는 `Authorization: Bearer`로 키를 포함해야 합니다.
- `HTTP_TIMEOUT_SEC` (기본 `10`) 체크 요청 전체 타임아웃
- `WEBHOOK_TIMEOUT_SEC` (기본 `5`) 알림 웹훅 타임아웃

### API 예시

체크 생성:

```bash
curl -X POST http://localhost:8080/checks \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "main-api",
    "url": "https://example.com/health",
    "method": "GET",
    "intervalSec": 30,
    "timeoutSec": 5,
    "expectedStatus": 200,
    "notifyOnRecovery": true,
    "enabled": true
  }'
```

체크 목록:

```bash
curl http://localhost:8080/checks
```

수동 실행:

```bash
curl -X POST http://localhost:8080/checks/<id>/run
```

### 상세 한글 가이드

자세한 사용법은 `USER_GUIDE_KO.md`를 확인하세요.

---

## English

UptimeGuardian is a Go service that periodically checks HTTP endpoints and sends webhook alerts when status changes occur (failure/recovery).

### Features

- CRUD API for checks
- Manual run endpoint (`POST /checks/{id}/run`)
- Per-check scheduler (`intervalSec`)
- Transition-based alerts (DOWN and optional RECOVERY)
- JSON file persistence

### Quick Start

```bash
cp .env.example .env
docker compose up -d --build
curl http://localhost:8080/health
```

### Environment Variables

- `PORT` (default: `8080`)
- `DATA_FILE` (default: `./data/checks.json`)
- `ALERT_WEBHOOK_URL` (leave empty to disable alerts)
- `REQUEST_USER_AGENT`
- `API_KEY` (optional). If set, requests must include `X-API-Key` or `Authorization: Bearer`.
- `HTTP_TIMEOUT_SEC` (default: `10`) overall HTTP check timeout
- `WEBHOOK_TIMEOUT_SEC` (default: `5`) webhook alert timeout

### API Example

Create a check:

```bash
curl -X POST http://localhost:8080/checks \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "main-api",
    "url": "https://example.com/health",
    "method": "GET",
    "intervalSec": 30,
    "timeoutSec": 5,
    "expectedStatus": 200,
    "notifyOnRecovery": true,
    "enabled": true
  }'
```

List checks:

```bash
curl http://localhost:8080/checks
```

Run check manually:

```bash
curl -X POST http://localhost:8080/checks/<id>/run
```
