#!/bin/bash

set -e

echo "📦 Installing Golang..."
sudo apt update && sudo apt install -y golang

echo "📁 Cloning or updating the repository..."
cd ~
if [ ! -d "kuccps" ]; then
    git clone https://github.com/excreal/kuccps.git
else
    cd kuccps
    git pull
    cd ..
fi

echo "🔨 Building the project..."

cd ~/kuccps/
bash build.sh
cd ~/kuccps/

echo "✅ Setup complete!"
