# UptimeGuardian 사용설명서 (한글)

UptimeGuardian은 서버/웹서비스의 상태를 주기적으로 확인하고, 장애 발생(또는 복구) 시 웹훅으로 알림을 보내는 프로그램입니다.

## 1) 준비 사항

- Docker 및 Docker Compose 사용 가능 환경
- 점검 대상 URL (예: `https://example.com/health`)
- 알림을 받고 싶은 웹훅 URL (선택)

## 2) 설치 및 실행

프로젝트 루트에서 아래 순서로 진행하세요.

```bash
cp .env.example .env
docker compose up -d --build
```

정상 실행 확인:

```bash
curl http://localhost:8080/health
```

응답 예시:

```json
{"status":"ok"}
```

## 3) 환경 변수 설정

`.env` 파일에서 값을 설정합니다.

- `PORT` : API 포트 (기본값 `8080`)
- `DATA_FILE` : 체크 정보 저장 파일 경로 (기본값 `./data/checks.json`)
- `ALERT_WEBHOOK_URL` : 장애/복구 알림 웹훅 URL
- `REQUEST_USER_AGENT` : 점검 요청 시 사용할 User-Agent

웹훅 알림이 필요 없으면 `ALERT_WEBHOOK_URL`을 비워두면 됩니다.

## 4) 기본 사용 흐름

1. 체크 대상 등록
2. 등록된 체크 목록 확인
3. 필요 시 수동 점검 실행
4. 주기 점검 및 알림 자동 동작

## 5) API 사용법

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

- `name`: 체크 이름
- `url`: 점검할 URL (`http` 또는 `https`)
- `method`: `GET` 또는 `HEAD`
- `intervalSec`: 점검 주기(초, 최소 5)
- `timeoutSec`: 요청 타임아웃(초, 최소 1)
- `expectedStatus`: 기대 HTTP 상태코드 (100~599)
- `notifyOnRecovery`: 복구 알림 전송 여부
- `enabled`: 점검 활성화 여부

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

## 6) 알림 동작 방식

- 이전 상태와 현재 상태가 달라질 때만 알림을 전송합니다.
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

## 7) 데이터 보관

- 체크 정보와 마지막 점검 결과는 `DATA_FILE` 경로(JSON)에 저장됩니다.
- `docker-compose.yml` 기본 설정에서는 `./data`가 컨테이너와 볼륨으로 연결됩니다.

## 8) 운영 팁

- 점검 주기는 서비스 성격에 맞게 설정하세요. (예: 30~60초)
- 타임아웃은 너무 길지 않게 유지하세요. (예: 3~5초)
- 장애 감지 정확도를 높이려면 헬스 전용 엔드포인트(`/health`)를 사용하세요.

## 9) 중지/재시작/로그 확인

```bash
docker compose stop
docker compose start
docker compose restart
docker compose logs -f
```
