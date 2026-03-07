# UptimeGuardian 사용설명서 (한글)

UptimeGuardian은 서버/웹서비스의 상태를 주기적으로 확인하고, 상태 변화(장애/복구) 발생 시 **웹훅으로 알림**을 보내는 간단한 업타임 모니터링 서비스입니다.

- HTTP(S) 엔드포인트를 `GET` 또는 `HEAD`로 점검
- 체크 대상은 CRUD API로 관리
- 상태 변화가 있을 때만 알림 전송 (DOWN, 선택적으로 RECOVERY)
- 체크 정보와 마지막 결과를 JSON 파일로 저장

---

## 1) 준비 사항

- Docker 및 Docker Compose 사용 가능 환경
- 점검 대상 URL (예: `https://example.com/health`)
- (선택) 알림을 받고 싶은 웹훅 URL

---

## 2) 설치 및 실행

프로젝트 루트에서 아래 순서로 진행하세요.

### 2-1. 환경파일 생성

```bash
cp .env.example .env
```

### 2-2. 실행

#### Docker Compose v2

```bash
docker compose up -d --build
```

#### Docker Compose v1

```bash
docker-compose up -d --build
```

### 2-3. 정상 실행 확인

```bash
curl http://localhost:8080/health
```

응답 예시:

```json
{"status":"ok"}
```

---

## 3) 환경 변수 설정

`.env` 파일에서 값을 설정합니다.

### 기본 설정

- `PORT` : API 포트 (기본값 `8080`)
- `DATA_FILE` : 체크 정보 저장 파일 경로 (기본값 `./data/checks.json`)
- `ALERT_WEBHOOK_URL` : 장애/복구 알림 웹훅 URL (비우면 알림 비활성)
- `REQUEST_USER_AGENT` : 점검 요청 시 사용할 User-Agent

### 3-1. API 인증(권장)

인터넷에 노출하거나, 같은 네트워크 내 다른 사람이 접근 가능할 때는 **반드시 API 인증을 켜는 것을 권장**합니다.

- `API_KEY` : 설정하면 모든 API 요청에 키가 필요합니다.
  - `X-API-Key: <API_KEY>`
  - 또는 `Authorization: Bearer <API_KEY>`

`API_KEY`가 비어 있으면(기본값) API는 인증 없이 열려 있습니다. (로컬 개발용)

### 3-2. 타임아웃

- `HTTP_TIMEOUT_SEC` (기본 10): 체크 요청 전체 타임아웃
- `WEBHOOK_TIMEOUT_SEC` (기본 5): 알림 웹훅 전송 타임아웃

---

## 4) 기본 사용 흐름

1. 체크 대상 등록
2. 등록된 체크 목록 확인
3. 필요 시 수동 점검 실행
4. 백그라운드에서 주기 점검 및 상태 변화 알림 자동 동작

---

## 5) API 사용법

기본 주소: `http://localhost:8080`

### 5-0. 인증 헤더 예시 (API_KEY 설정했을 때)

```bash
curl -H 'X-API-Key: <API_KEY>' http://localhost:8080/checks
# 또는
curl -H 'Authorization: Bearer <API_KEY>' http://localhost:8080/checks
```

### 5-1. 체크 등록

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

필드 설명:

- `name`: 체크 이름 (필수)
- `url`: 점검할 URL (`http` 또는 `https`, 필수)
- `method`: `GET` 또는 `HEAD` (기본 `GET`)
- `intervalSec`: 점검 주기(초, 최소 5, 기본 60)
- `timeoutSec`: 요청 타임아웃(초, 최소 1, 기본 5)
- `expectedStatus`: 기대 HTTP 상태코드 (100~599, 기본 200)
- `notifyOnRecovery`: 복구 알림 전송 여부 (기본 false)
- `enabled`: 점검 활성화 여부 (기본 true)

주의:
- JSON 바디는 1MB로 제한됩니다.
- 정의되지 않은 필드를 보내면(오타 등) `400` 에러로 거절됩니다.

### 5-2. 체크 목록 조회

```bash
curl http://localhost:8080/checks
```

### 5-3. 특정 체크 조회

```bash
curl http://localhost:8080/checks/<id>
```

### 5-4. 체크 수정

```bash
curl -X PUT http://localhost:8080/checks/<id> \
  -H 'Content-Type: application/json' \
  -d '{"intervalSec":60,"enabled":true}'
```

### 5-5. 체크 삭제

```bash
curl -X DELETE http://localhost:8080/checks/<id>
```

### 5-6. 수동 점검 실행

```bash
curl -X POST http://localhost:8080/checks/<id>/run
```

---

## 6) 알림(웹훅) 동작 방식

- **이전 상태와 현재 상태가 달라질 때만** 알림을 전송합니다.
- DOWN 전환 시에는 항상 알림 전송 시도
- UP(복구) 전환 시에는 `notifyOnRecovery=true`인 체크만 알림 전송

웹훅 페이로드 예시:

```json
{
  "checkId": "abc123",
  "checkName": "main-api",
  "url": "https://example.com/health",
  "previousStatus": "up",
  "currentStatus": "down",
  "statusCode": 500,
  "latencyMs": 88,
  "checkedAt": "2026-03-07T21:00:00Z",
  "error": ""
}
```

### Discord/Slack으로 알림 보내기(중요)

UptimeGuardian은 **일반 JSON 페이로드**를 웹훅으로 전송합니다.
하지만 Discord/Slack의 Incoming Webhook은 보통 `{ "content": ... }`(Discord) 또는 `{ "text": ... }`(Slack) 같은 **자기들 전용 포맷**을 기대합니다.

그래서 보통은 아래 중 하나로 구성합니다.

1) **웹훅 릴레이(추천)**
- `ALERT_WEBHOOK_URL`을 "내가 만든 중간 서버"(Cloudflare Worker/AWS Lambda/간단한 서버)로 지정
- 중간 서버가 UptimeGuardian payload를 받아 Discord/Slack 포맷으로 변환해 전송

2) **자동화 툴 사용**
- n8n / Zapier / Make.com: JSON 수신 → 변환 → Discord/Slack 전송

#### 가장 단순한 릴레이 예시(Node.js)

`relay.js` 파일을 만들고:

```js
import express from "express";

const app = express();
app.use(express.json({ limit: "1mb" }));

const DISCORD_WEBHOOK_URL = process.env.DISCORD_WEBHOOK_URL;

app.post("/uptimeguardian", async (req, res) => {
  const p = req.body;
  const text = `[${p.currentStatus?.toUpperCase?.() ?? ""}] ${p.checkName}\n${p.url}\nstatus=${p.statusCode} latency=${p.latencyMs}ms`;

  if (DISCORD_WEBHOOK_URL) {
    await fetch(DISCORD_WEBHOOK_URL, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ content: text }),
    });
  }

  res.json({ ok: true });
});

app.listen(3000, () => console.log("relay listening on :3000"));
```

실행:

```bash
npm i express
node relay.js
```

그리고 UptimeGuardian에는 아래처럼 설정합니다:

- `ALERT_WEBHOOK_URL=http://<relay-host>:3000/uptimeguardian`

운영 환경에서는 릴레이도 반드시 인증/HTTPS를 적용하는 것을 권장합니다.

---

## 7) 데이터 보관

- 체크 정보와 마지막 점검 결과는 `DATA_FILE` 경로(JSON)에 저장됩니다.
- `docker-compose.yml` 기본 설정에서는 `./data`가 컨테이너와 볼륨으로 연결됩니다.

---

## 8) 운영 팁 / 보안 팁

- 점검 주기는 서비스 성격에 맞게 설정하세요. (예: 30~60초)
- 타임아웃은 너무 길지 않게 유지하세요. (예: 3~5초)
- 장애 감지 정확도를 높이려면 헬스 전용 엔드포인트(`/health`)를 사용하세요.
- 외부에 노출한다면:
  - `API_KEY`를 꼭 설정하세요.
  - 가능하면 방화벽/리버스프록시로 접근을 제한하세요.

---

## 9) 중지/재시작/로그 확인

### Docker Compose v2

```bash
docker compose stop
docker compose start
docker compose restart
docker compose logs -f
```

### Docker Compose v1

```bash
docker-compose stop
docker-compose start
docker-compose restart
docker-compose logs -f
```
