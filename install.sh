#!/bin/bash

# 获取系统类型
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# 获取 CPU 架构
ARCH=$(uname -m)

# 根据系统类型和 CPU 架构选择不同的 curl 下载地址
if [[ "$OS" == "linux" ]]; then
  if [[ "$ARCH" == "amd64" ]]; then
    REALARCH="amd64"
  elif [[ "$ARCH" == "arm64" ]]; then
    REALARCH="arm64"
  elif [[ "$ARCH" == "x86_64" ]]; then
    REALARCH="amd64"  
  else
    echo "Unsupported architecture: $ARCH"
    exit 1
  fi
elif [[ "$OS" == "darwin" ]]; then
   if [[ "$ARCH" == "amd64" ]]; then
        REALARCH="amd64"
    elif [[ "$ARCH" == "arm64" ]]; then
        REALARCH="arm64"
    elif [[ "$ARCH" == "x86_64" ]]; then
        REALARCH="amd64"    
    else 
        echo "Unsupported architecture: $ARCH"
        exit 1
    fi
else
  echo "Unsupported operating system: $OS"
  exit 1
fi

echo "dolowadning wukongimcli for $OS/$REALARCH ..."

# 下载 curl
curl -L "https://github.com/WuKongIM/WuKongIMCli/releases/download/$1/wukongimcli-$OS-$REALARCH" -o /usr/local/bin/wk
chmod +x /usr/local/bin/wk