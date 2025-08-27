#!/bin/bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ANDROID_DIR="${REPO_ROOT}/android"
GO_SRC_DIR="${REPO_ROOT}/android/go/mobile"
JNI_LIBS_DIR="${ANDROID_DIR}/app/src/main/jniLibs"
FRONTEND_DIR="${REPO_ROOT}/frontend"

# 用法说明
usage() {
  echo "用法: $0 [--skip-frontend]" >&2
  echo "环境变量: SKIP_FRONTEND=1 跳过前端构建" >&2
}

SKIP_FRONTEND_FLAG=0
for arg in "${@:-}"; do
  case "$arg" in
    --skip-frontend) SKIP_FRONTEND_FLAG=1 ;;
    -h|--help) usage; exit 0 ;;
    *) ;;
  esac
done

: "${ANDROID_NDK_HOME:?请先设置 ANDROID_NDK_HOME 指向 Android NDK 根目录}"

API=21
# 自动检测 prebuilt 目录（darwin-x86_64 或 darwin-arm64）
TOOLCHAIN_PREBUILT_DIR=$(ls -d "${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/"* 2>/dev/null | head -n 1 || true)
if [[ -z "${TOOLCHAIN_PREBUILT_DIR}" || ! -d "${TOOLCHAIN_PREBUILT_DIR}" ]]; then
  echo "未找到 NDK 预编译工具链目录: ${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/*" >&2
  exit 1
fi

build_one() {
  local ABI="$1" GOARCH="$2" TRIPLE="$3" EXTRA_ENV="$4"
  local OUT_DIR="${JNI_LIBS_DIR}/${ABI}"
  mkdir -p "${OUT_DIR}"

  export CGO_ENABLED=1
  export GOOS=android
  export GOARCH="${GOARCH}"
  export CC="${TOOLCHAIN_PREBUILT_DIR}/bin/${TRIPLE}${API}-clang"
  # 兼容 armeabi-v7a
  if [[ -n "${EXTRA_ENV}" ]]; then
    eval "${EXTRA_ENV}"
  fi

  echo "→ 构建 ${ABI} (${GOARCH}) 使用 ${CC}"
  ( cd "${GO_SRC_DIR}" && go build -buildmode=c-shared -o "${OUT_DIR}/libgo_video_crawler.so" )
}

# 前置：构建前端（供 go:embed 与静态资源使用）
if [[ "${SKIP_FRONTEND_FLAG}" -eq 0 && "${SKIP_FRONTEND:-0}" -ne 1 ]]; then
  echo "[go_lib] 构建前端..."
  cd "${FRONTEND_DIR}"
  if [ -f package-lock.json ]; then
    npm ci --no-audit --no-fund
  else
    npm install --no-audit --no-fund
  fi
  npm run build
else
  echo "[go_lib] 跳过前端构建 (通过 --skip-frontend 或 SKIP_FRONTEND=1)"
fi

# 清理旧产物
echo "清理旧的 JNI 库文件..."
rm -rf "${JNI_LIBS_DIR}"
mkdir -p "${JNI_LIBS_DIR}"

# 进入仓库根确保依赖可解析
cd "${REPO_ROOT}"
go mod tidy

# 依次构建 4 个 ABI
build_one arm64-v8a arm64 aarch64-linux-android ""
build_one armeabi-v7a arm armv7a-linux-androideabi "export GOARM=7"
build_one x86 386 i686-linux-android ""
build_one x86_64 amd64 x86_64-linux-android ""

echo "生成完成: ${JNI_LIBS_DIR}/{arm64-v8a,armeabi-v7a,x86,x86_64}/libgo_video_crawler.so"
