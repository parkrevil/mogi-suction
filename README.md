# Mogi Suction - Go Monorepo

간단한 Hello World 서버와 클라이언트를 포함한 Go 모노레포입니다.

## 프로젝트 구조

```
mogi-suction/
├── apps/
│   ├── server/          # HTTP 서버
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── .air.toml   # Hot reload 설정
│   └── client/          # HTTP 클라이언트
│       ├── main.go
│       ├── go.mod
│       └── .air.toml   # Hot reload 설정
├── bin/                 # 빌드된 바이너리 (자동 생성)
├── go.mod              # 루트 모듈
├── Makefile            # 빌드 및 실행 스크립트
└── README.md
```

## 기능

### Server (`apps/server`)
- 간단한 HTTP 서버
- Hello World 응답
- 포트 8080에서 실행

### Client (`apps/client`)
- 패킷 캡처 클라이언트
- gopacket을 사용한 TCP 패킷 분석
- 포트 16000의 TCP 트래픽 캡처
- 이더넷, IP, TCP 레이어 분석
- 페이로드 내용 출력 (UTF-8 및 Hex)

## 시작하기

### 1. 의존성 설치
```bash
make deps
```

### 2. 서버 실행
```bash
make run-server
```
또는 개발 모드 (hot reload):
```bash
make watch-server
```

### 3. 클라이언트 실행
```bash
# 일반 실행
make run-client

# 개발 모드 (hot reload)
make watch-client

# 패킷 캡처 실행 (sudo 권한 필요)
sudo ./bin/client

# 개발 모드 (sudo 권한으로 hot reload)
make watch-client-sudo
```

## 빌드

### 전체 빌드
```bash
make build
```

### 개별 빌드
```bash
make build-server  # 서버만 빌드
make build-client  # 클라이언트만 빌드
```

## 기능

### 패킷 캡처
- **포트**: TCP 포트 16000
- **필터**: BPF 필터 `tcp port 16000`
- **분석**: 이더넷, IP, TCP 레이어
- **출력**: MAC 주소, IP 주소, 포트, TCP 플래그, 페이로드

### HTTP 서버
- `GET /` - Hello World 응답

## 개발 환경

### Hot Reload (Air)
프로젝트는 [Air](https://github.com/air-verse/air)를 사용하여 hot reload 기능을 제공합니다.

```bash
# 서버 hot reload
make watch-server

# 클라이언트 hot reload  
make watch-client
```

## 사용 가능한 명령어

```bash
make help              # 도움말 보기
make deps              # 의존성 설치
make build             # 전체 빌드
make build-server      # 서버 빌드
make build-client      # 클라이언트 빌드
make run-server        # 서버 실행
make run-client        # 클라이언트 실행
make watch-server          # 서버 hot reload
make watch-client          # 클라이언트 hot reload
make watch-client-sudo     # 클라이언트 hot reload (sudo)
make test              # 테스트 실행
make clean             # 빌드 아티팩트 정리
```

## 개발 환경

- Go 1.24+
- Air (hot reload)
- 네트워크 연결

## 라이센스

MIT License 